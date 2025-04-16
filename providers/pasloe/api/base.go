package api

import (
	"context"
	"github.com/Fesaa/Media-Provider/db/models"
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

type IDAble interface {
	ID() string
	Label() string
}

func NewDownloadableFromBlock[T IDAble](scope *dig.Scope, handler string, block DownloadInfoProvider[T]) *DownloadBase[T] {
	var base = &DownloadBase[T]{}

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest,
		client Client,
		log zerolog.Logger,
		signalR services.SignalRService,
		notification services.NotificationService,
		transLoco services.TranslocoService,
		preferences models.Preferences,
		fs afero.Afero,
	) {
		base = &DownloadBase[T]{
			infoProvider: block,
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
		}
	}))

	return base
}

type Content struct {
	Name string
	Path string
}

type DownloadBase[T IDAble] struct {
	infoProvider DownloadInfoProvider[T]

	Client       Client
	Log          zerolog.Logger
	contentState payload.ContentState
	SignalR      services.SignalRService
	Notifier     services.NotificationService
	TransLoco    services.TranslocoService
	fs           afero.Afero

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

	preferences models.Preferences
	Preference  *models.Preference

	ContentDownloaded int
	ImagesDownloaded  int
	LastTime          time.Time
	LastRead          int

	failedDownloads int

	cancel context.CancelFunc
	Wg     *sync.WaitGroup
}

func (d *DownloadBase[T]) FailedDownloads() int {
	return d.failedDownloads
}

func (d *DownloadBase[T]) Request() payload.DownloadRequest {
	return d.Req
}

func (d *DownloadBase[T]) SetState(state payload.ContentState) {
	d.contentState = state
	d.SignalR.StateUpdate(d.id, d.contentState)
}

func (d *DownloadBase[T]) Id() string {
	return d.id
}

func (d *DownloadBase[T]) GetBaseDir() string {
	return d.baseDir
}

func (d *DownloadBase[T]) GetDownloadDir() string {
	title := d.infoProvider.Title()
	if title == "" {
		return d.baseDir
	}
	return path.Join(d.baseDir, title)
}

func (d *DownloadBase[T]) GetOnDiskContent() []Content {
	return d.ExistingContent
}

func (d *DownloadBase[T]) ExistingContentNames() []string {
	return utils.Map(d.ExistingContent, func(t Content) string {
		return t.Name
	})
}

func (d *DownloadBase[T]) GetContentByName(name string) (Content, bool) {
	for _, content := range d.ExistingContent {
		if content.Name == name {
			return content, true
		}
	}
	return Content{}, false
}

func (d *DownloadBase[T]) GetContentByPath(path string) (Content, bool) {
	for _, content := range d.ExistingContent {
		if content.Path == path {
			return content, true
		}
	}
	return Content{}, false
}

func (d *DownloadBase[T]) GetNewContentNamed() []string {
	return utils.Map(d.ToDownload, func(t T) string {
		return t.Label()
	})
}

func (d *DownloadBase[T]) GetNewContent() []string {
	return d.HasDownloaded
}

func (d *DownloadBase[T]) GetToRemoveContent() []string {
	return d.ToRemoveContent
}

func (d *DownloadBase[T]) GetInfo() payload.InfoStat {
	return payload.InfoStat{
		Provider:     models.DYNASTY,
		Id:           d.Id(),
		ContentState: d.contentState,
		Name:         d.infoProvider.Title(),
		RefUrl:       d.infoProvider.RefUrl(),
		Size:         strconv.Itoa(d.Size()) + " Chapters",
		Downloading:  d.State() == payload.ContentStateDownloading,
		Progress:     utils.Percent(int64(d.ContentDownloaded), int64(d.Size())),
		SpeedType:    payload.IMAGES,
		Speed:        d.Speed(),
		DownloadDir:  d.GetDownloadDir(),
	}
}

// Cancel calls d.cancel and send a StopRequest with DeleteFiles=true to the Client
func (d *DownloadBase[T]) Cancel() {
	d.Log.Trace().Msg("calling cancel on content")
	if d.cancel != nil {
		d.cancel()
	}
	if d.Client.Content(d.id) != nil {
		if err := d.Client.RemoveDownload(payload.StopRequest{
			Provider:    d.infoProvider.Provider(),
			Id:          d.id,
			DeleteFiles: true,
		}); err != nil {
			d.Log.Warn().Err(err).Msg("failed to cancel download")
		}
	}
}

func (d *DownloadBase[T]) StartLoadInfo() {
	if d.cancel != nil {
		d.Log.Debug().Msg("content already started")
		return
	}

	loadInfoStart := time.Now()

	d.SetState(payload.ContentStateLoading)
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel
	d.Log.Debug().Msg("loading content info")

	p, err := d.preferences.GetComplete()
	if err != nil {
		d.Log.Error().Err(err).Msg("unable to get preferences, some features may not work")
	}
	d.Preference = p

	start := time.Now()
	select {
	case <-ctx.Done():
		return
	case <-d.infoProvider.LoadInfo(ctx):
	}

	elapsed := time.Since(start)

	d.Log = d.Log.With().Str("title", d.infoProvider.Title()).Logger()
	d.Log.Debug().Dur("elapsed", elapsed).Msg("Content has downloaded all information")

	start = time.Now()
	d.checkContentOnDisk()
	data := d.infoProvider.All()
	d.ToDownload = utils.Filter(data, func(t T) bool {
		download := d.infoProvider.ShouldDownload(t)
		if !download {
			d.Log.Trace().Str("key", d.infoProvider.ContentKey(t)).Msg("content already downloaded, skipping")
		} else {
			d.Log.Trace().Str("key", d.infoProvider.ContentKey(t)).Msg("adding content to download queue")
		}
		return download
	})

	elapsed = time.Since(start)
	if elapsed > time.Second*5 {
		d.Log.Warn().Dur("elapsed", elapsed).Msg("checking which content must be downloaded took a long time")

		if d.Req.IsSubscription {
			d.Notifier.NotifyContent(
				d.TransLoco.GetTranslation("warn"),
				d.TransLoco.GetTranslation("long-on-disk-check", d.infoProvider.Title()),
				d.TransLoco.GetTranslation("long-on-disk-check-body", elapsed),
				models.Orange)
		}
	}

	if len(d.ToDownload) == 0 {
		d.Log.Debug().Msg("no chapters to download, stopping")

		d.SetState(payload.ContentStateWaiting)

		req := payload.StopRequest{
			Provider:    d.Req.Provider,
			Id:          d.Id(),
			DeleteFiles: false,
		}
		if err = d.Client.RemoveDownload(req); err != nil {
			d.Log.Error().Err(err).Msg("error while cleaning up")
		}
		return
	}

	d.SetState(utils.Ternary(d.Req.DownloadMetadata.StartImmediately,
		payload.ContentStateReady,
		payload.ContentStateWaiting))
	d.SignalR.UpdateContentInfo(d.GetInfo())

	d.Log.Debug().Int("all", len(data)).Int("filtered", len(d.ToDownload)).
		Dur("StartLoadInfo#duration", time.Since(loadInfoStart)).Msg("downloaded content filtered")
}

func (d *DownloadBase[T]) StartDownload() {
	if d.State() != payload.ContentStateReady && d.State() != payload.ContentStateWaiting {
		d.Log.Warn().Any("state", d.State()).Msg("cannot start download, content not ready")
		return
	}
	d.SetState(payload.ContentStateDownloading)
	go d.startDownload()
}

func (d *DownloadBase[T]) State() payload.ContentState {
	return d.contentState
}

// Speed returns the speed at which this content is downloading images
func (d *DownloadBase[T]) Speed() int64 {
	if d.contentState != payload.ContentStateDownloading {
		return 0
	}
	diff := d.ImagesDownloaded - d.LastRead
	timeDiff := max(time.Since(d.LastTime).Seconds(), 1)

	d.LastRead = d.ImagesDownloaded
	d.LastTime = time.Now()
	return int64(math.Ceil(float64(diff) / timeDiff))
}

func (d *DownloadBase[T]) Size() int {
	if len(d.ToDownloadUserSelected) == 0 {
		return len(d.ToDownload)
	}

	return len(d.ToDownloadUserSelected)
}

func (d *DownloadBase[T]) UpdateProgress() {
	d.SignalR.ProgressUpdate(payload.ContentProgressUpdate{
		ContentId: d.id,
		Progress:  utils.Percent(int64(d.ContentDownloaded), int64(d.Size())),
		SpeedType: payload.IMAGES,
		Speed:     utils.Ternary(d.State() != payload.ContentStateCleanup, d.Speed(), 0),
	})
}
