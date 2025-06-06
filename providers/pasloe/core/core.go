package core

import (
	"context"
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
	"strconv"
	"sync"
	"time"
)

func New[T Chapter](scope *dig.Scope, handler string, provider DownloadInfoProvider[T]) *Core[T] {
	var base *Core[T]

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest,
		client Client,
		log zerolog.Logger,
		signalR services.SignalRService,
		notification services.NotificationService,
		transLoco services.TranslocoService,
		preferences models.Preferences,
		fs afero.Afero,
		httpClient *menou.Client,
	) {
		base = &Core[T]{
			infoProvider: provider,
			Client:       client,
			Log:          log.With().Str("handler", handler).Str("id", req.Id).Logger(),
			id:           req.Id,
			baseDir:      req.BaseDir,
			TempTitle:    req.TempTitle,
			maxImages:    min(client.GetConfig().GetMaxConcurrentImages(), 4),
			Req:          req,
			LastTime:     time.Now(),
			contentState: payload.ContentStateQueued,
			SignalR:      signalR,
			Notifier:     notification,
			TransLoco:    transLoco,
			preferences:  preferences,
			fs:           fs,
			httpClient:   httpClient,
		}
	}))

	return base
}

type Content struct {
	Name string
	Path string
}

type Core[T Chapter] struct {
	infoProvider DownloadInfoProvider[T]

	Client       Client
	Log          zerolog.Logger
	contentState payload.ContentState
	SignalR      services.SignalRService
	Notifier     services.NotificationService
	TransLoco    services.TranslocoService
	fs           afero.Afero
	httpClient   *menou.Client

	id        string
	baseDir   string
	TempTitle string
	maxImages int
	Req       payload.DownloadRequest

	ToDownload      []T
	HasDownloaded   []string
	ExistingContent []Content
	ToRemoveContent []string

	// ToDownloadUserSelected are the ids of the content selected by the user to download in the UI
	ToDownloadUserSelected []string

	preferences   models.Preferences
	Preference    *models.Preference
	hasWarnedTags bool

	ContentDownloaded int
	ImagesDownloaded  int
	LastTime          time.Time
	LastRead          int

	failedDownloads int

	cancel context.CancelFunc
	Wg     *sync.WaitGroup
}

func (c *Core[T]) DisplayInformation() DisplayInformation {
	return DisplayInformation{
		Name: func() string {
			if c.Req.IsSubscription && c.Req.Sub != nil {
				return c.Req.Sub.Info.Title
			}
			return c.infoProvider.Title()
		}(),
	}
}

func (c *Core[T]) FailedDownloads() int {
	return c.failedDownloads
}

func (c *Core[T]) Request() payload.DownloadRequest {
	return c.Req
}

func (c *Core[T]) SetState(state payload.ContentState) {
	c.contentState = state
	c.SignalR.StateUpdate(c.id, c.contentState)
}

func (c *Core[T]) Id() string {
	return c.id
}

func (c *Core[T]) GetBaseDir() string {
	return c.baseDir
}

func (c *Core[T]) GetDownloadDir() string {
	title := c.infoProvider.Title()
	if title == "" {
		return c.baseDir
	}
	return path.Join(c.baseDir, title)
}

func (c *Core[T]) GetOnDiskContent() []Content {
	return c.ExistingContent
}

func (c *Core[T]) ExistingContentNames() []string {
	return utils.Map(c.ExistingContent, func(t Content) string {
		return t.Name
	})
}

func (c *Core[T]) GetContentByName(name string) (Content, bool) {
	for _, content := range c.ExistingContent {
		if content.Name == name {
			return content, true
		}
	}
	return Content{}, false
}

func (c *Core[T]) GetContentByPath(path string) (Content, bool) {
	for _, content := range c.ExistingContent {
		if content.Path == path {
			return content, true
		}
	}
	return Content{}, false
}

func (c *Core[T]) GetNewContentNamed() []string {
	return utils.Map(c.ToDownload, func(t T) string {
		return t.Label()
	})
}

func (c *Core[T]) GetNewContent() []string {
	return c.HasDownloaded
}

func (c *Core[T]) GetToRemoveContent() []string {
	return c.ToRemoveContent
}

func (c *Core[T]) GetInfo() payload.InfoStat {
	return payload.InfoStat{
		Provider:     models.DYNASTY,
		Id:           c.Id(),
		ContentState: c.contentState,
		Name:         c.infoProvider.Title(),
		RefUrl:       c.infoProvider.RefUrl(),
		Size:         strconv.Itoa(c.Size()) + " Chapters",
		Downloading:  c.State() == payload.ContentStateDownloading,
		Progress:     utils.Percent(int64(c.ContentDownloaded), int64(c.Size())),
		SpeedType:    payload.IMAGES,
		Speed:        c.Speed(),
		DownloadDir:  c.GetDownloadDir(),
	}
}

// Cancel calls d.cancel and send a StopRequest with DeleteFiles=true to the Client
func (c *Core[T]) Cancel() {
	c.Log.Trace().Msg("calling cancel on content")
	if c.cancel != nil {
		c.cancel()
	}
	if c.Client.Content(c.id) != nil {
		if err := c.Client.RemoveDownload(payload.StopRequest{
			Provider:    c.infoProvider.Provider(),
			Id:          c.id,
			DeleteFiles: true,
		}); err != nil {
			c.Log.Warn().Err(err).Msg("failed to cancel download")
		}
	}
}

func (c *Core[T]) initializeLoadInfo() (context.Context, bool) {
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

func (c *Core[T]) loadContentInfo(ctx context.Context) bool {
	start := time.Now()
	select {
	case <-ctx.Done():
		return false
	case <-c.infoProvider.LoadInfo(ctx):
	}

	elapsed := time.Since(start)

	c.Log = c.Log.With().Str("title", c.infoProvider.Title()).Logger()
	c.Log.Debug().Dur("elapsed", elapsed).Msg("Content has downloaded all information")
	return true
}

// prepareContentToDownload checks what content exists on disk and filters what needs to be downloaded
func (c *Core[T]) prepareContentToDownload() ([]T, time.Duration) {
	start := time.Now()
	c.loadContentOnDisk()

	data := c.infoProvider.All()
	c.ToDownload = utils.Filter(data, func(t T) bool {
		download := c.infoProvider.ShouldDownload(t)
		if !download {
			c.Log.Trace().Str("key", c.infoProvider.ContentKey(t)).Msg("content already downloaded, skipping")
		} else {
			c.Log.Trace().Str("key", c.infoProvider.ContentKey(t)).Msg("adding content to download queue")
		}
		return download
	})

	return data, time.Since(start)
}

func (c *Core[T]) handleLongDiskCheck(elapsed time.Duration) {
	if elapsed > time.Second*5 {
		c.Log.Warn().Dur("elapsed", elapsed).Msg("checking which content must be downloaded took a long time")

		if c.Req.IsSubscription {
			c.Notifier.NotifyContent(
				c.TransLoco.GetTranslation("warn"),
				c.TransLoco.GetTranslation("long-on-disk-check", c.infoProvider.Title()),
				c.TransLoco.GetTranslation("long-on-disk-check-body", elapsed),
				models.Orange)
		}
	}
}

func (c *Core[T]) handleNoContentToDownload() bool {
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

func (c *Core[T]) finalizeLoadInfo(data []T, loadInfoStart time.Time) {
	c.SetState(utils.Ternary(c.Req.DownloadMetadata.StartImmediately,
		payload.ContentStateReady,
		payload.ContentStateWaiting))
	c.SignalR.UpdateContentInfo(c.GetInfo())

	c.Log.Debug().Int("all", len(data)).Int("filtered", len(c.ToDownload)).
		Dur("StartLoadInfo#duration", time.Since(loadInfoStart)).Msg("downloaded content filtered")
}

func (c *Core[T]) StartLoadInfo() {
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

func (c *Core[T]) StartDownload() {
	if c.State() != payload.ContentStateReady && c.State() != payload.ContentStateWaiting {
		c.Log.Warn().Any("state", c.State()).Msg("cannot start download, content not ready")
		return
	}
	c.SetState(payload.ContentStateDownloading)
	go c.startDownload()
}

func (c *Core[T]) State() payload.ContentState {
	return c.contentState
}

// Speed returns the speed at which this content is downloading images
func (c *Core[T]) Speed() int64 {
	if c.contentState != payload.ContentStateDownloading {
		return 0
	}
	diff := c.ImagesDownloaded - c.LastRead
	timeDiff := max(time.Since(c.LastTime).Seconds(), 1)

	c.LastRead = c.ImagesDownloaded
	c.LastTime = time.Now()
	return int64(math.Ceil(float64(diff) / timeDiff))
}

func (c *Core[T]) Size() int {
	if len(c.ToDownloadUserSelected) == 0 {
		return len(c.ToDownload)
	}

	return len(c.ToDownloadUserSelected)
}

func (c *Core[T]) UpdateProgress() {
	c.SignalR.ProgressUpdate(payload.ContentProgressUpdate{
		ContentId: c.id,
		Progress:  utils.Percent(int64(c.ContentDownloaded), int64(c.Size())),
		SpeedType: payload.IMAGES,
		Speed:     utils.Ternary(c.State() != payload.ContentStateCleanup, c.Speed(), 0),
	})
}
