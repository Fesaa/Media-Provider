package core

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

func (c *Core[C, S]) abortDownload(reason error) {
	if errors.Is(reason, context.Canceled) {
		return
	}

	c.Log.Error().Err(reason).Msg("error while downloading content; cleaning up")

	if c.cancel != nil {
		c.cancel()
	}

	// wait for all download tasks to finish
	c.wg.Wait()
	c.IoWg.Wait()

	req := payload.StopRequest{
		Provider:    c.Req.Provider,
		Id:          c.Id(),
		DeleteFiles: true,
	}
	if err := c.Client.RemoveDownload(req); err != nil {
		c.Log.Error().Err(err).Msg("error while cleaning up")
	}
	c.Notifier.Notify(context.TODO(), models.Notification{
		Title:   "Failed download",
		Summary: fmt.Sprintf("%s failed to download", c.impl.Title()),
		Body:    fmt.Sprintf("Download failed for %s, because %v", c.impl.Title(), reason),
		Colour:  models.Error,
		Group:   models.GroupError,
	})
}

func (c *Core[C, S]) filterContentByUserSelection() {
	if len(c.ToDownloadUserSelected) == 0 {
		return
	}

	currentSize := len(c.ToDownload)
	c.ToDownload = utils.Filter(c.ToDownload, func(t C) bool {
		return slices.Contains(c.ToDownloadUserSelected, t.GetId())
	})
	c.Log.Debug().Int("size", currentSize).Int("newSize", len(c.ToDownload)).
		Msg("content further filtered after user has made a selection in the UI")

	if len(c.ToRemoveContent) == 0 {
		return
	}

	paths := utils.Map(c.ToDownload, func(t C) string {
		return c.ContentPath(t) + ".cbz"
	})
	c.ToRemoveContent = utils.Filter(c.ToRemoveContent, func(s string) bool {
		return slices.Contains(paths, s)
	})
}

func (c *Core[C, S]) processDownloads(ctx context.Context, wg *sync.WaitGroup) error {
	for _, content := range c.ToDownload {
		select {
		case <-ctx.Done():
			return errors.New("download cancelled")
		default:
			wg.Add(1)
			err := func() error {
				defer wg.Done()
				return c.downloadContent(ctx, content)
			}()
			if err != nil {
				c.abortDownload(err)
				return err
			}
		}
		c.UpdateProgress()
	}
	return nil
}

func (c *Core[C, S]) cleanupAfterDownload(wg *sync.WaitGroup) {
	wg.Wait()
	req := payload.StopRequest{
		Provider:    c.Req.Provider,
		Id:          c.Id(),
		DeleteFiles: false,
	}
	if err := c.Client.RemoveDownload(req); err != nil {
		c.Log.Error().Err(err).Msg("error while cleaning up files")
	}
}

func (c *Core[C, S]) startDownload(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	c.cancel = cancel

	data := c.GetAllLoadedChapters()
	c.Log.Debug().Int("size", len(data)).Msg("downloading content")

	c.filterContentByUserSelection()

	c.Log.Info().
		Int("all", len(data)).
		Int("toDownload", len(c.ToDownload)).
		Int("reDownloads", len(c.ToRemoveContent)).
		Str("into", c.GetDownloadDir()).
		Msg("downloading content")

	start := time.Now()

	c.wg = &sync.WaitGroup{}
	c.IoWg = &sync.WaitGroup{}
	c.IOWorkCh = make(chan IOTask, c.maxImages*2) // Allow for some buffer as I/O may be slower than downloading (webp)

	c.startProgressUpdater(ctx)
	c.StartIOWorkers(ctx)

	if err := c.processDownloads(ctx, c.wg); err != nil {
		c.Log.Trace().Err(err).Msg("download failed")
		close(c.IOWorkCh)
		return
	}

	c.Log.Debug().Dur("elapsed", time.Since(start)).
		Msg("All content has been downloaded, waiting for I/O workers to finish")

	close(c.IOWorkCh)
	c.SetState(payload.ContentStateCleanup)
	c.IoWg.Wait()

	c.Log.Info().Dur("elapsed", time.Since(start)).Msg("Finished downloading content")

	c.cleanupAfterDownload(c.wg)
}

// downloadContent handles the full download of one chapter
func (c *Core[C, S]) downloadContent(parentCtx context.Context, chapter C) error {
	dCtx, err := c.constructDownloadContext(parentCtx, chapter)
	if err != nil || dCtx == nil {
		return err
	}

	defer dCtx.Cancel()

	contentPath := c.ContentPath(chapter)
	if err = c.fs.MkdirAll(contentPath, 0755); err != nil {
		return err
	}

	// Mark as downloaded as soon as the directory is created as we need to remove it in case of an error
	c.HasDownloaded = append(c.HasDownloaded, contentPath)

	if err = c.impl.WriteContentMetaData(dCtx.Ctx, chapter); err != nil { //nolint: contextcheck
		c.Log.Warn().Err(err).Msg("error writing metadata")
	}

	dCtx.log.Debug().Int("size", len(dCtx.Urls)).Msg("downloading images")
	start := time.Now()

	go dCtx.ProduceUrls()

	// Reset image progress
	atomic.StoreInt64(&c.ImagesDownloaded, 0)
	atomic.StoreInt64(&c.LastRead, 0)
	c.TotalChapterImages = len(dCtx.Urls)

	dCtx.StartDownloadWorkers()

	dCtx.DownloadWg.Wait()
	dCtx.log.Debug().Dur("elapsed", time.Since(start)).Msg("Finished downloading all remote content for chapter")

	select {
	case err = <-dCtx.ErrCh:
		return err
	default:
	}

	if len(dCtx.Urls) < 5 {
		time.Sleep(1 * time.Second)
	}

	c.ContentDownloaded++
	return nil
}

// constructDownloadContext load the required information to start downloading the chapter, and then returns the context
// with all fields set
func (c *Core[C, S]) constructDownloadContext(ctx context.Context, chapter C) (*DownloadContext[C, S], error) {
	downloadContext, cancel := context.WithCancel(ctx)

	dCtx := DownloadContext[C, S]{
		Ctx:     downloadContext,
		Core:    c,
		Chapter: chapter,
		Cancel:  cancel,

		RateLimiter: rate.NewLimiter(rate.Limit(c.maxImages), c.maxImages),
		DownloadWg:  &sync.WaitGroup{},
		DownloadCh:  make(chan DownloadTask, c.maxImages),

		ErrCh: make(chan error, 1),

		log: c.ContentLogger(chapter),
	}

	dCtx.log.Trace().Msg("loading content info and creating directories")

	urls, err := c.impl.ContentUrls(ctx, chapter)
	if err != nil {
		return &dCtx, err
	}
	if len(urls) == 0 {
		dCtx.log.Warn().Msg("content has no downloadable urls? Unexpected? Report this!")
		return nil, nil
	}

	dCtx.Urls = urls
	return &dCtx, nil
}

// DownloadContext convince wrapper around the download of a chapter
type DownloadContext[C Chapter, S Series[C]] struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	Core    *Core[C, S]
	Chapter C
	Urls    []string

	RateLimiter *rate.Limiter
	DownloadWg  *sync.WaitGroup
	DownloadCh  chan DownloadTask

	ErrCh chan error

	log zerolog.Logger
}

type DownloadTask struct {
	idx int
	url string
}

type IOTask struct {
	data  []byte
	path  string
	dTask DownloadTask
}

// CancelWithError queue the given error into the ErrCh, and call cancel if it succeeded. Otherwise, do nothing
// does not block
func (d *DownloadContext[C, S]) CancelWithError(err error) {
	d.log.Trace().Err(err).Msg("Canceling download task")
	select {
	case d.ErrCh <- err:
		d.Cancel()
	default:
	}
}

// IsCancelled checks if the context has been cancelled
// Returns true if cancelled, false otherwise
func (d *DownloadContext[C, S]) IsCancelled() bool {
	select {
	case <-d.Ctx.Done():
		return true
	default:
		return false
	}
}

// ProduceUrls queues a DownloadTask for each URL and closes the channel after
func (d *DownloadContext[C, S]) ProduceUrls() {
	defer close(d.DownloadCh)

	d.log.Trace().Int("urls", len(d.Urls)).Msg("queuing urls to download")

	for idx, url := range d.Urls {
		if d.IsCancelled() {
			return
		}

		select {
		case d.DownloadCh <- DownloadTask{idx + 1, url}:
		case <-d.Ctx.Done():
			return
		}
	}
}

// StartDownloadWorkers starts cap(DownloadCh) DownloadWorker threads
func (d *DownloadContext[C, S]) StartDownloadWorkers() {
	for worker := range cap(d.DownloadCh) {
		d.DownloadWg.Add(1)
		go d.DownloadWorker(fmt.Sprintf("DownloadWorker#%d", worker))
	}
}

// DownloadWorker reads from DownloadCh; will download the remote content and queue a IO task
// has its own internal retry system. After one retry fail with stop content download
func (d *DownloadContext[C, S]) DownloadWorker(id string) {
	defer d.DownloadWg.Done()

	log := d.log.With().Str("DownloadWorker#", id).Logger()

	failedTasks := d.processDownloads(log, d.DownloadCh, false)

	if len(failedTasks) == 0 {
		return
	}

	d.log.Debug().Int("failedDownloads", len(failedTasks)).
		Msg("Some images failed to download, retrying")

	failedCh := make(chan DownloadTask, len(failedTasks))
	for _, task := range failedTasks {
		failedCh <- task
	}
	close(failedCh)

	// We can ignore the return value, errors on retry stop the download
	d.processDownloads(log, failedCh, true)
}

// processDownloads tries downloading tasks in the channel and sends them into the IO Ch. If isRetry is false will
// return failed tasks, otherwise stops download
func (d *DownloadContext[C, S]) processDownloads(log zerolog.Logger, taskCh <-chan DownloadTask, isRetry bool) []DownloadTask {
	failedTasks := make([]DownloadTask, 0)

	for task := range taskCh {
		if d.IsCancelled() {
			return failedTasks
		}

		if err := d.RateLimiter.Wait(d.Ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("rate limiter wait failed")
			}

			return failedTasks
		}

		log.Trace().Int("idx", task.idx).Str("url", task.url).Bool("isRetry", isRetry).
			Msg("downloading page")

		data, err := d.Core.Download(d.Ctx, task.url)
		if err != nil {
			if d.IsCancelled() {
				return failedTasks
			}

			if isRetry {
				log.Error().Err(err).Int("idx", task.idx).Str("url", task.url).
					Msg("retry download failed, ending content download")

				d.CancelWithError(fmt.Errorf("final download failed on url %s; %w", task.url, err))
				return failedTasks
			}

			failedTasks = append(failedTasks, task)
			atomic.AddInt64(&d.Core.failedDownloads, 1)

			log.Warn().Err(err).Int("idx", task.idx).Str("url", task.url).
				Msg("download has failed for a page for the first time, trying page again at the end")
			continue
		}

		atomic.AddInt64(&d.Core.ImagesDownloaded, 1)
		if d.IsCancelled() {
			return failedTasks
		}

		select {
		case d.Core.IOWorkCh <- IOTask{data, d.Core.ContentPath(d.Chapter), task}:
		case <-d.Ctx.Done():
			return failedTasks
		}
	}

	return failedTasks
}

// startProgressUpdater start a goroutine sending payload.EventTypeContentProgressUpdate every 2s for this chapter
func (c *Core[C, S]) startProgressUpdater(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.UpdateProgress()
			}
		}
	}()
}
