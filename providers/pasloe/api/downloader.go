package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"math"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
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

func (d *DownloadBase[T]) Message(msg payload.Message) (payload.Message, error) {
	var jsonBytes []byte
	var err error
	switch msg.MessageType {
	case payload.MessageListContent:
		jsonBytes, err = json.Marshal(d.infoProvider.ContentList())
	case payload.SetToDownload:
		err = d.SetUserFiltered(msg.Data)
	case payload.StartDownload:
		err = d.MarkReady()
	default:
		return payload.Message{}, services.ErrUnknownMessageType
	}

	if err != nil {
		return payload.Message{}, err
	}

	return payload.Message{
		Provider:    d.Req.Provider,
		ContentId:   d.id,
		MessageType: msg.MessageType,
		Data:        jsonBytes,
	}, nil
}

func (d *DownloadBase[T]) MarkReady() error {
	if d.contentState != payload.ContentStateWaiting {
		return services.ErrWrongState
	}

	if d.Client.CanStart(d.Req.Provider) {
		go d.StartDownload()
		return nil
	}

	d.SetState(payload.ContentStateReady)
	return nil
}

func (d *DownloadBase[T]) SetUserFiltered(msg json.RawMessage) error {
	if d.contentState != payload.ContentStateWaiting &&
		d.contentState != payload.ContentStateReady {
		return services.ErrWrongState
	}

	var filter []string
	err := json.Unmarshal(msg, &filter)
	if err != nil {
		return err
	}
	d.ToDownloadUserSelected = filter
	d.SignalR.SizeUpdate(d.Id(), strconv.Itoa(d.Size())+" Chapters")
	return nil
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

func (d *DownloadBase[T]) checkContentOnDisk() {
	d.Log.Debug().Str("dir", d.GetDownloadDir()).Msg("checking content on disk")
	content, err := d.readDirectoryForContent(d.GetDownloadDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d.Log.Trace().Msg("directory not found, fresh download")
		} else {
			d.Log.Warn().Err(err).Msg("unable to check for already downloaded content. Downloading all")
		}
		d.ExistingContent = []Content{}
		return
	}

	d.Log.Trace().Str("content", fmt.Sprintf("%v", content)).Msg("found following content on disk")
	d.ExistingContent = content
}

func (d *DownloadBase[T]) readDirectoryForContent(p string) ([]Content, error) {
	entries, err := d.fs.ReadDir(path.Join(d.Client.GetBaseDir(), p))
	if err != nil {
		return nil, err
	}

	out := make([]Content, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirContent, err2 := d.readDirectoryForContent(path.Join(p, entry.Name()))
			if err2 != nil {
				return nil, err2
			}
			out = append(out, dirContent...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".cbz") {
			d.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue

		}

		matches := d.infoProvider.IsContent(entry.Name())
		if !matches {
			d.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue
		}
		d.Log.Trace().Str("file", entry.Name()).Msg("found  content on disk")
		out = append(out, Content{
			Name: entry.Name(),
			Path: path.Join(p, entry.Name()),
		})
	}

	return out, nil
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

func (d *DownloadBase[T]) abortDownload(reason error) {
	if errors.Is(reason, context.Canceled) {
		return
	}

	d.Log.Error().Err(reason).Msg("error while downloading content; cleaning up")
	req := payload.StopRequest{
		Provider:    d.Req.Provider,
		Id:          d.Id(),
		DeleteFiles: true,
	}
	if err := d.Client.RemoveDownload(req); err != nil {
		d.Log.Error().Err(err).Msg("error while cleaning up")
	}
	d.Notifier.Notify(models.Notification{
		Title:   "Failed download",
		Summary: fmt.Sprintf("%s failed to download", d.infoProvider.Title()),
		Body:    fmt.Sprintf("Download failed for %s, because %v", d.infoProvider.Title(), reason),
		Colour:  models.Red,
		Group:   models.GroupError,
	})
}

func (d *DownloadBase[T]) startDownload() {
	// Overwrite cancel, as we're doing something else
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel

	data := d.infoProvider.All()
	d.Log.Trace().Int("size", len(data)).Msg("downloading content")
	d.Wg = &sync.WaitGroup{}

	if len(d.ToDownloadUserSelected) > 0 {
		currentSize := len(d.ToDownload)
		d.ToDownload = utils.Filter(d.ToDownload, func(t T) bool {
			return slices.Contains(d.ToDownloadUserSelected, t.ID())
		})
		d.Log.Debug().Int("size", currentSize).Int("newSize", len(d.ToDownload)).
			Msg("content further filtered after user has made a selection in the UI")

		if len(d.ToRemoveContent) > 0 {
			paths := utils.Map(d.ToDownload, func(t T) string {
				return d.infoProvider.ContentPath(t) + ".cbz"
			})
			d.ToRemoveContent = utils.Filter(d.ToRemoveContent, func(s string) bool {
				return slices.Contains(paths, s)
			})
		}

	}

	d.Log.Info().
		Int("all", len(data)).
		Int("toDownload", len(d.ToDownload)).
		Int("reDownloads", len(d.ToRemoveContent)).
		Str("into", d.GetDownloadDir()).
		Msg("downloading content")
	for _, content := range d.ToDownload {
		select {
		case <-ctx.Done():
			d.Wg.Wait()
			return
		default:
			d.Wg.Add(1)
			err := d.downloadContent(ctx, content)
			d.Wg.Done()
			if err != nil {
				d.abortDownload(err)
				d.Wg.Wait()
				return
			}
		}
		d.UpdateProgress()
	}

	d.Wg.Wait()
	req := payload.StopRequest{
		Provider:    d.Req.Provider,
		Id:          d.Id(),
		DeleteFiles: false,
	}
	if err := d.Client.RemoveDownload(req); err != nil {
		d.Log.Error().Err(err).Msg("error while cleaning up files")
	}
}

type downloadUrl struct {
	idx int
	url string
}

func (d *DownloadBase[T]) downloadContent(ctx context.Context, t T) error {
	l := d.infoProvider.ContentLogger(t)
	l.Trace().Msg("downloading content")

	contentPath := d.infoProvider.ContentPath(t)
	if err := d.fs.MkdirAll(contentPath, 0755); err != nil {
		return err
	}
	d.HasDownloaded = append(d.HasDownloaded, contentPath)

	urls, err := d.infoProvider.ContentUrls(ctx, t)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		l.Warn().Msg("content has no downloadable urls? Unexpected? Report this!")
		return nil
	}

	if err = d.infoProvider.WriteContentMetaData(t); err != nil {
		d.Log.Warn().Err(err).Msg("error writing meta data")
	}

	l.Debug().Int("size", len(urls)).Msg("downloading images")

	urlCh := make(chan downloadUrl, d.maxImages)
	errCh := make(chan error, 1)
	defer close(errCh)

	wg := &sync.WaitGroup{}
	innerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer close(urlCh)
		for i, url := range urls {
			select {
			case <-ctx.Done():
				return
			case <-innerCtx.Done():
				return
			case urlCh <- downloadUrl{url: url, idx: i + 1}:
			}
		}
	}()

	go func() {
		for range time.Tick(2 * time.Second) {
			select {
			case <-innerCtx.Done():
				return
			case <-ctx.Done():
				return
			default:
				d.UpdateProgress()
			}
		}
	}()

	for range d.maxImages {
		wg.Add(1)
		go d.channelConsumer(innerCtx, cancel, ctx, t, l, urlCh, errCh, wg)
	}

	wg.Wait()

	select {
	case err = <-errCh:
		return err
	default:
	}

	if len(urls) < 5 {
		time.Sleep(1 * time.Second)
	}

	d.ContentDownloaded++
	return nil
}

func (d *DownloadBase[T]) channelConsumer(
	innerCtx context.Context,
	cancel context.CancelFunc,
	ctx context.Context,
	t T,
	l zerolog.Logger,
	urlCh chan downloadUrl,
	errCh chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	failedCh := make(chan downloadUrl)

	for urlData := range urlCh {
		select {
		case <-innerCtx.Done():
			return
		case <-ctx.Done():
			return
		default:
			l.Trace().Int("idx", urlData.idx).Str("url", urlData.url).Msg("downloading page")

			err := d.infoProvider.DownloadContent(urlData.idx, t, urlData.url)
			if err == nil {
				time.Sleep(1 * time.Second)
				continue
			}

			select {
			case <-innerCtx.Done():
				return
			case <-ctx.Done():
				return
			case failedCh <- urlData:
				d.failedDownloads++
				l.Warn().Err(err).Int("idx", urlData.idx).Str("url", urlData.url).
					Msg("download has failed for a page for the first time, trying page again at the end")
			}

			time.Sleep(1 * time.Second)
		}
	}

	close(failedCh)
	for reTry := range failedCh {
		if err := d.infoProvider.DownloadContent(reTry.idx, t, reTry.url); err != nil {
			l.Error().Err(err).Str("url", reTry.url).Msg("Failed final download")
			select {
			case errCh <- fmt.Errorf("final download failed %w", err):
				cancel()
			default:
			}
			return
		}
		d.failedDownloads++
		time.Sleep(1 * time.Second)
	}
}
