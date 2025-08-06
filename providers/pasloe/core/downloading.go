package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"path"
	"slices"
	"sync"
	"time"
)

func (c *Core[C, S]) abortDownload(reason error) {
	if errors.Is(reason, context.Canceled) {
		return
	}

	c.Log.Error().Err(reason).Msg("error while downloading content; cleaning up")
	req := payload.StopRequest{
		Provider:    c.Req.Provider,
		Id:          c.Id(),
		DeleteFiles: true,
	}
	if err := c.Client.RemoveDownload(req); err != nil {
		c.Log.Error().Err(err).Msg("error while cleaning up")
	}
	c.Notifier.Notify(models.Notification{
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
			wg.Wait()
			return errors.New("download cancelled")
		default:
			wg.Add(1)
			err := c.downloadContent(ctx, content)
			wg.Done()
			if err != nil {
				c.abortDownload(err)
				wg.Wait()
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

func (c *Core[C, S]) startDownload() {
	// Overwrite cancel, as we're doing something else
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	data := c.GetAllLoadedChapters()
	c.Log.Trace().Int("size", len(data)).Msg("downloading content")

	c.filterContentByUserSelection()

	c.Log.Info().
		Int("all", len(data)).
		Int("toDownload", len(c.ToDownload)).
		Int("reDownloads", len(c.ToRemoveContent)).
		Str("into", c.GetDownloadDir()).
		Msg("downloading content")

	start := time.Now()
	wg := &sync.WaitGroup{}

	if err := c.processDownloads(ctx, wg); err != nil {
		c.Log.Trace().Err(err).Msg("download failed")
		return
	}

	c.Log.Info().Dur("elapsed", time.Since(start)).Msg("Finished downloading content")

	c.cleanupAfterDownload(wg)
}

// downloadContent handles the full download of one chapter
func (c *Core[C, S]) downloadContent(ctx context.Context, chapter C) error {
	dCtx, err := c.ConstructDownloadContext(ctx, chapter)
	if err != nil {
		return err
	}

	if dCtx == nil {
		return nil
	}

	defer dCtx.Cancel()

	contentPath := c.ContentPath(chapter)
	if err = c.fs.MkdirAll(contentPath, 0755); err != nil {
		return err
	}
	c.HasDownloaded = append(c.HasDownloaded, contentPath)

	if err = c.impl.WriteContentMetaData(ctx, chapter); err != nil {
		c.Log.Warn().Err(err).Msg("error writing meta data")
	}

	dCtx.log.Debug().Int("size", len(dCtx.Urls)).Msg("downloading images")
	start := time.Now()

	c.startProgressUpdater(ctx, dCtx.Ctx)

	go dCtx.ProduceUrls()

	dCtx.StartDownloadWorkers()
	dCtx.StartIOWorkers()

	dCtx.DownloadWg.Wait()
	dCtx.log.Debug().Dur("elapsed", time.Since(start)).Msg("Finished downloading all remote content, waiting for I/O goroutines to finish")
	close(dCtx.IOCh)

	dCtx.IoWg.Wait()
	dCtx.log.Debug().Dur("elapsed", time.Since(start)).Msg("All I/O goroutines finished, checking for errors and cleaning up")

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

// ConstructDownloadContext load the required information to start downloading the chapter, and then returns the context
// with all fields set
func (c *Core[C, S]) ConstructDownloadContext(ctx context.Context, chapter C) (*DownloadContext[C, S], error) {
	downloadContext, cancel := context.WithCancel(context.Background())

	dCtx := DownloadContext[C, S]{
		GlobalCtx: ctx,
		Ctx:       downloadContext,
		Core:      c,
		Chapter:   chapter,
		Cancel:    cancel,

		DownloadWg: &sync.WaitGroup{},
		DownloadCh: make(chan DownloadTask, c.maxImages),

		IoWg: &sync.WaitGroup{},
		IOCh: make(chan IOTask, c.maxImages*2), // Allow for some buffer as I/O may be slower than downloading (webp)

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
	GlobalCtx context.Context
	Ctx       context.Context
	Cancel    context.CancelFunc

	Core    *Core[C, S]
	Chapter C
	Urls    []string

	DownloadWg *sync.WaitGroup
	DownloadCh chan DownloadTask

	IoWg *sync.WaitGroup
	IOCh chan IOTask

	ErrCh chan error

	log zerolog.Logger
}

type DownloadTask struct {
	idx int
	url string
}

type IOTask struct {
	data  []byte
	dTask DownloadTask
}

// CancelWithError queue the given error into the ErrCh, and call cancel if it succeeded. Otherwise, do nothing
// does not block
func (d *DownloadContext[C, S]) CancelWithError(err error) {
	select {
	case d.ErrCh <- err:
		d.Cancel()
	default:
	}
}

// ProduceUrls queues a DownloadTask for each URL and closes the channel after
func (d *DownloadContext[C, S]) ProduceUrls() {
	defer close(d.DownloadCh)

	d.log.Trace().Int("urls", len(d.Urls)).Msg("queuing urls to download")

	for idx, url := range d.Urls {
		select {
		case <-d.GlobalCtx.Done():
			return
		case <-d.Ctx.Done():
			return
		case d.DownloadCh <- DownloadTask{idx + 1, url}:
		}
	}
}

// StartDownloadWorkers starts cap(DownloadCh) DownloadWorker threads
func (d *DownloadContext[C, S]) StartDownloadWorkers() {
	d.log.Trace().Int("workers", cap(d.DownloadCh)).Msg("starting download workers")

	for worker := range cap(d.DownloadCh) {
		d.DownloadWg.Add(1)
		go d.DownloadWorker(fmt.Sprintf("DownloadWorker#%d", worker))
	}
}

// StartIOWorkers starts cap(IOCh) IOWorker threads
func (d *DownloadContext[C, S]) StartIOWorkers() {
	for worker := range cap(d.IOCh) {
		d.IoWg.Add(1)
		go d.IOWorker(fmt.Sprintf("IOWorker#%d", worker))
	}
}

// DownloadWorker reads from DownloadCh; will download the remote content and queue a IO task
// has its own internal retry system. After one retry fail with stop content download
func (d *DownloadContext[C, S]) DownloadWorker(id string) {
	log := d.log.With().Str("worker", id).Logger()
	defer d.DownloadWg.Done()

	failedDownloadCh := make(chan DownloadTask, len(d.Urls))

	rateLimiter := time.NewTicker(time.Second)
	defer rateLimiter.Stop()

	for task := range d.DownloadCh {
		select {
		case <-d.Ctx.Done():
			return
		case <-d.GlobalCtx.Done():
			return
		case <-rateLimiter.C:
		}

		log.Trace().Int("idx", task.idx).Str("url", task.url).Msg("downloading page")

		data, err := d.Core.Download(d.Ctx, task.url)
		if err == nil {
			d.Core.ImagesDownloaded++
			d.IOCh <- IOTask{data, task}
			continue
		}

		select {
		case <-d.GlobalCtx.Done():
			return
		case <-d.Ctx.Done():
			return
		case failedDownloadCh <- task:
			d.Core.failedDownloads++
			log.Warn().Err(err).Int("idx", task.idx).Str("url", task.url).
				Msg("download has failed for a page for the first time, trying page again at the end")
		}
	}

	close(failedDownloadCh)

	if len(failedDownloadCh) == 0 {
		return
	}

	for task := range failedDownloadCh {
		select {
		case <-d.Ctx.Done():
			return
		case <-d.GlobalCtx.Done():
			return
		case <-rateLimiter.C:
		}

		log.Trace().Int("idx", task.idx).Str("url", task.url).Msg("downloading page")

		data, err := d.Core.Download(d.Ctx, task.url)
		if err == nil {
			d.IOCh <- IOTask{data, task}
			continue
		}

		select {
		case <-d.GlobalCtx.Done():
			return
		case <-d.Ctx.Done():
			return
		default:
			log.Error().Err(err).Int("idx", task.idx).Str("url", task.url).
				Msg("retry download failed, ending content download")
			d.CancelWithError(fmt.Errorf("final download failed on url %s; %w", task.url, err))
			return
		}
	}
}

// IOWorker reads from IOCh; converts to webp and writes to disk
func (d *DownloadContext[C, S]) IOWorker(id string) {
	log := d.log.With().Str("worker", id).Logger()

	defer d.IoWg.Done()

	for task := range d.IOCh {
		data, ok := d.Core.imageService.ConvertToWebp(task.data)

		ext := utils.Ternary(ok, ".webp", utils.Ext(task.dTask.url))
		filePath := path.Join(d.Core.ContentPath(d.Chapter), fmt.Sprintf("page %s"+ext, utils.PadInt(task.dTask.idx, 4)))

		if err := d.Core.fs.WriteFile(filePath, data, 0755); err != nil {
			log.Error().Err(err).Msg("error writing file")
			d.CancelWithError(fmt.Errorf("error writing file %s: %w", filePath, err))
		}
	}

}

// startProgressUpdater start a goroutine sending payload.EventTypeContentProgressUpdate every 2s for this chapter
func (c *Core[C, S]) startProgressUpdater(ctx context.Context, innerCtx context.Context) {
	go func() {
		for range time.Tick(2 * time.Second) {
			select {
			case <-innerCtx.Done():
				return
			case <-ctx.Done():
				return
			default:
				c.UpdateProgress()
			}
		}
	}()
}
