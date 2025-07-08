package core

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

func New[C Chapter, S Series[C]](scope *dig.Scope, handler string, provider DownloadInfoProvider[C]) *Core[C, S] {
	var base *Core[C, S]

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest,
		client Client,
		log zerolog.Logger,
		signalR services.SignalRService,
		notification services.NotificationService,
		transLoco services.TranslocoService,
		preferences models.Preferences,
		archiveService services.ArchiveService,
		imageService services.ImageService,
		fs afero.Afero,
		httpClient *menou.Client,
		settingsService services.SettingsService,
	) error {

		settings, err := settingsService.GetSettingsDto()
		if err != nil {
			return err
		}

		base = &Core[C, S]{
			impl:           provider,
			Client:         client,
			Log:            log.With().Str("handler", handler).Str("id", req.Id).Logger(),
			id:             req.Id,
			baseDir:        req.BaseDir,
			maxImages:      utils.Clamp(settings.MaxConcurrentImages, 1, 5),
			Req:            req,
			LastTime:       time.Now(),
			contentState:   payload.ContentStateQueued,
			SignalR:        signalR,
			Notifier:       notification,
			TransLoco:      transLoco,
			archiveService: archiveService,
			imageService:   imageService,
			preferences:    preferences,
			fs:             fs,
			httpClient:     httpClient,
		}

		return nil
	}))

	return base
}

type Content struct {
	Name string
	Path string
	// The following fields are parsed from the file name
	Chapter string
	Volume  string
}

type Core[C Chapter, S Series[C]] struct {
	impl DownloadInfoProvider[C]

	Client         Client
	Log            zerolog.Logger
	contentState   payload.ContentState
	SignalR        services.SignalRService
	Notifier       services.NotificationService
	TransLoco      services.TranslocoService
	archiveService services.ArchiveService
	imageService   services.ImageService
	fs             afero.Afero
	httpClient     *menou.Client

	id        string
	baseDir   string
	maxImages int
	Req       payload.DownloadRequest

	SeriesInfo S
	// hasDuplicatedChapters is true if the same chapter number is used across different volumes
	// forcing us to use volumes in the file name
	hasDuplicatedChapters utils.Settable[bool]

	ToDownload []C
	// Path to the directory container the chapters files
	HasDownloaded []string
	// Content already on disk before download started
	ExistingContent []Content
	// Content on disk that has to be removed as it has been redownloaded
	ToRemoveContent []string

	// ToDownloadUserSelected are the ids of the content selected by the user to download in the UI
	ToDownloadUserSelected []string

	preferences   models.Preferences
	Preference    *models.Preference
	hasWarnedTags bool

	// # Chapters downloaded
	ContentDownloaded int
	// Total amount of images downloaded
	ImagesDownloaded int
	LastTime         time.Time
	LastRead         int

	failedDownloads int

	cancel context.CancelFunc
}

func (c *Core[C, S]) DisplayInformation() DisplayInformation {
	return DisplayInformation{
		Name: func() string {
			if c.Req.IsSubscription && c.Req.Sub != nil {
				return c.Req.Sub.Info.Title
			}
			return c.impl.Title()
		}(),
	}
}

func (c *Core[C, S]) FailedDownloads() int {
	return c.failedDownloads
}

func (c *Core[C, S]) Request() payload.DownloadRequest {
	return c.Req
}

func (c *Core[C, S]) SetState(state payload.ContentState) {
	c.contentState = state
	c.SignalR.StateUpdate(c.id, c.contentState)
}

func (c *Core[C, S]) Id() string {
	return c.id
}

func (c *Core[C, S]) GetBaseDir() string {
	return c.baseDir
}

func (c *Core[C, S]) GetDownloadDir() string {
	title := c.impl.Title()
	if title == "" {
		return c.baseDir
	}
	return path.Join(c.baseDir, title)
}

func (c *Core[C, S]) GetOnDiskContent() []Content {
	return c.ExistingContent
}

func (c *Core[C, S]) ExistingContentNames() []string {
	return utils.Map(c.ExistingContent, func(t Content) string {
		return t.Name
	})
}

func (c *Core[C, S]) GetContentByName(name string) (Content, bool) {
	for _, content := range c.ExistingContent {
		if content.Name == name {
			return content, true
		}
	}
	return Content{}, false
}

func (c *Core[C, S]) GetContentByPath(path string) (Content, bool) {
	for _, content := range c.ExistingContent {
		if content.Path == path {
			return content, true
		}
	}
	return Content{}, false
}

func (c *Core[C, S]) GetContentByVolumeAndChapter(volume string, chapter string) (Content, bool) {
	for _, content := range c.ExistingContent {
		if content.Volume == volume && content.Chapter == chapter {
			return content, true
		}
		if content.Volume == "" && content.Chapter == chapter {
			return content, true
		}
	}
	return Content{}, false
}

func (c *Core[C, S]) GetNewContentNamed() []string {
	return utils.Map(c.ToDownload, func(t C) string {
		return t.Label()
	})
}

func (c *Core[C, S]) GetNewContent() []string {
	return c.HasDownloaded
}

func (c *Core[C, S]) GetToRemoveContent() []string {
	return c.ToRemoveContent
}

func (c *Core[C, S]) WillBeDownloaded(chapter C) bool {
	if len(c.ToDownloadUserSelected) > 0 {
		return slices.Contains(c.ToDownloadUserSelected, chapter.GetId())
	}

	return utils.Find(c.ToDownload, func(c C) bool {
		return c.GetId() == chapter.GetId()
	}) != nil
}

func (c *Core[C, S]) ContentList() []payload.ListContentData {
	chapters := c.GetAllLoadedChapters()
	if len(chapters) == 0 {
		return []payload.ListContentData{}
	}

	data := utils.GroupBy(chapters, func(v C) string {
		return v.GetVolume()
	})

	childrenFunc := func(chapters []C) []payload.ListContentData {
		slices.SortFunc(chapters, func(a, b C) int {
			if a.GetVolume() != b.GetVolume() {
				return (int)(utils.SafeFloat(b.GetVolume()) - utils.SafeFloat(a.GetVolume()))
			}
			return (int)(utils.SafeFloat(b.GetChapter()) - utils.SafeFloat(a.GetChapter()))
		})

		return utils.Map(chapters, func(chapter C) payload.ListContentData {
			return payload.ListContentData{
				SubContentId: chapter.GetId(),
				Selected:     c.WillBeDownloaded(chapter),
				Label:        strings.TrimSpace(chapter.GetTitle() + " " + chapter.Label()),
			}
		})
	}

	sortSlice := utils.Keys(data)
	slices.SortFunc(sortSlice, utils.SortFloats)

	out := make([]payload.ListContentData, 0, len(data))
	for _, volume := range sortSlice {
		chaptersInVolume := data[volume]

		// Do not add No Volume label if there are no volumes
		if volume == "" && len(sortSlice) == 1 {
			out = append(out, childrenFunc(chaptersInVolume)...)
			continue
		}

		out = append(out, payload.ListContentData{
			Label:    utils.Ternary(volume == "", "No Volume", fmt.Sprintf("Volume %s", volume)),
			Children: childrenFunc(chaptersInVolume),
		})
	}
	return out
}

func (c *Core[C, S]) GetInfo() payload.InfoStat {
	return payload.InfoStat{
		Provider:     models.DYNASTY,
		Id:           c.Id(),
		ContentState: c.contentState,
		Name:         c.impl.Title(),
		RefUrl:       c.impl.RefUrl(),
		Size:         strconv.Itoa(c.Size()) + " Chapters",
		Downloading:  c.State() == payload.ContentStateDownloading,
		Progress:     utils.Percent(int64(c.ContentDownloaded), int64(c.Size())),
		SpeedType:    payload.IMAGES,
		Speed:        c.Speed(),
		DownloadDir:  c.GetDownloadDir(),
	}
}

// Cancel calls d.cancel and send a StopRequest with DeleteFiles=true to the Client
func (c *Core[C, S]) Cancel() {
	c.Log.Trace().Msg("calling cancel on content")
	if c.cancel != nil {
		c.cancel()
	}
	if c.Client.Content(c.id) != nil {
		if err := c.Client.RemoveDownload(payload.StopRequest{
			Provider:    c.impl.Provider(),
			Id:          c.id,
			DeleteFiles: true,
		}); err != nil {
			c.Log.Warn().Err(err).Msg("failed to cancel download")
		}
	}
}

func (c *Core[C, S]) initializeLoadInfo() (context.Context, bool) {
	if c.cancel != nil {
		c.Log.Debug().Msg("content already started")
		return nil, false
	}

	c.SetState(payload.ContentStateLoading)
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.Log.Debug().Msg("loading content info")

	return ctx, true
}

func (c *Core[C, S]) loadContentInfo(ctx context.Context) bool {
	start := time.Now()
	select {
	case <-ctx.Done():
		return false
	case <-c.impl.LoadInfo(ctx):
	}

	elapsed := time.Since(start)

	c.Log = c.Log.With().Str("title", c.impl.Title()).Logger()
	c.Log.Debug().Dur("elapsed", elapsed).Msg("Content has downloaded all information")
	return true
}

// prepareContentToDownload checks what content exists on disk and filters what needs to be downloaded
func (c *Core[C, S]) prepareContentToDownload() ([]C, time.Duration) {
	start := time.Now()
	c.loadContentOnDisk()

	data := c.GetAllLoadedChapters()
	c.ToDownload = utils.Filter(data, c.ShouldDownload)
	return data, time.Since(start)
}

func (c *Core[C, S]) handleLongDiskCheck(elapsed time.Duration) {
	if elapsed > time.Second*5 {
		c.Log.Warn().Dur("elapsed", elapsed).Msg("checking which content must be downloaded took a long time")

		if c.Req.IsSubscription {
			c.Notifier.NotifyContent(
				c.TransLoco.GetTranslation("warn"),
				c.TransLoco.GetTranslation("long-on-disk-check", c.impl.Title()),
				c.TransLoco.GetTranslation("long-on-disk-check-body", elapsed),
				models.Orange)
		}
	}
}

func (c *Core[C, S]) handleNoContentToDownload() bool {
	if len(c.ToDownload) == 0 {
		c.Log.Debug().Msg("no chapters to download, stopping")

		c.SetState(payload.ContentStateWaiting)

		req := payload.StopRequest{
			Provider:    c.Req.Provider,
			Id:          c.Id(),
			DeleteFiles: false,
		}
		if err := c.Client.RemoveDownload(req); err != nil {
			c.Log.Error().Err(err).Msg("error while cleaning up")
		}
		return true
	}
	return false
}

func (c *Core[C, S]) finalizeLoadInfo(data []C, loadInfoStart time.Time) {
	c.SetState(utils.Ternary(c.Req.DownloadMetadata.StartImmediately,
		payload.ContentStateReady,
		payload.ContentStateWaiting))
	c.SignalR.UpdateContentInfo(c.GetInfo())

	c.Log.Debug().Int("all", len(data)).Int("filtered", len(c.ToDownload)).
		Dur("StartLoadInfo#duration", time.Since(loadInfoStart)).Msg("downloaded content filtered")
}

func (c *Core[C, S]) StartLoadInfo() {
	ctx, shouldContinue := c.initializeLoadInfo()
	if !shouldContinue {
		return
	}

	loadInfoStart := time.Now()

	p, err := c.preferences.GetComplete()
	if err != nil {
		c.Log.Error().Err(err).Msg("unable to get preferences, some features may not work")
	}
	c.Preference = p

	if !c.loadContentInfo(ctx) {
		return
	}

	data, elapsed := c.prepareContentToDownload()
	c.handleLongDiskCheck(elapsed)

	if c.handleNoContentToDownload() {
		return
	}

	c.finalizeLoadInfo(data, loadInfoStart)
}

func (c *Core[C, S]) StartDownload() {
	if c.State() != payload.ContentStateReady && c.State() != payload.ContentStateWaiting {
		c.Log.Warn().Any("state", c.State()).Msg("cannot start download, content not ready")
		return
	}
	c.SetState(payload.ContentStateDownloading)
	go c.startDownload()
}

func (c *Core[C, S]) State() payload.ContentState {
	return c.contentState
}

// Speed returns the speed at which this content is downloading images
func (c *Core[C, S]) Speed() int64 {
	if c.contentState != payload.ContentStateDownloading {
		return 0
	}
	diff := c.ImagesDownloaded - c.LastRead
	timeDiff := max(time.Since(c.LastTime).Seconds(), 1)

	c.LastRead = c.ImagesDownloaded
	c.LastTime = time.Now()
	return int64(math.Ceil(float64(diff) / timeDiff))
}

func (c *Core[C, S]) Size() int {
	if len(c.ToDownloadUserSelected) == 0 {
		return len(c.ToDownload)
	}

	return len(c.ToDownloadUserSelected)
}

func (c *Core[C, S]) UpdateProgress() {
	c.SignalR.ProgressUpdate(payload.ContentProgressUpdate{
		ContentId: c.id,
		Progress:  utils.Percent(int64(c.ContentDownloaded), int64(c.Size())),
		SpeedType: payload.IMAGES,
		Speed:     utils.Ternary(c.State() != payload.ContentStateCleanup, c.Speed(), 0),
	})
}
