package publication

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
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
)

// pipeline convince wrapper around the download of a chapter
type pipeline struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	Publication *publication
	Chapter     Chapter
	Urls        []string

	RateLimiter *rate.Limiter
	DownloadWg  *sync.WaitGroup
	DownloadCh  chan downloadTask

	ErrCh chan error

	log zerolog.Logger
}

func (pl *pipeline) isCancelled() bool {
	select {
	case <-pl.Ctx.Done():
		return true
	default:
		return false
	}
}

func (pl *pipeline) ProduceUrls() {
	defer close(pl.DownloadCh)

	for idx, url := range pl.Urls {
		if pl.isCancelled() {
			return
		}

		select {
		case pl.DownloadCh <- downloadTask{idx + 1, url}:
		case <-pl.Ctx.Done():
			return
		}
	}
}

func (pl *pipeline) StartDownloadWorkers() {
	for worker := range cap(pl.DownloadCh) {
		pl.DownloadWg.Add(1)
		go pl.DownloadWorker(fmt.Sprintf("DownloadWorker#%d", worker))
	}
}

func (pl *pipeline) DownloadWorker(id string) {
	defer pl.DownloadWg.Done()

	ctx, span := tracing.TracerPasloe.Start(pl.Ctx, tracing.SpanPasloeDownloadWorker)
	defer span.End()

	span.SetAttributes(attribute.String("worker.id", id))

	log := pl.log.With().Str("DownloadWorker#", id).Logger()

	failedTasks := pl.processDownloads(ctx, log, pl.DownloadCh, false)
	if len(failedTasks) == 0 {
		return
	}

	pl.log.Debug().Int("failedTasks", len(failedTasks)).
		Msg("some tasks failed to complete, retrying")

	failedCh := make(chan downloadTask, len(failedTasks))
	for _, task := range failedTasks {
		failedCh <- task
	}
	close(failedCh)
	pl.processDownloads(ctx, log, failedCh, true)
}

func (pl *pipeline) processDownloads(ctx context.Context, log zerolog.Logger, ch chan downloadTask, isRetry bool) []downloadTask {
	span := trace.SpanFromContext(ctx)
	failedTasks := make([]downloadTask, 0)

	for task := range ch {
		if pl.isCancelled() {
			return failedTasks
		}

		if err := pl.RateLimiter.Wait(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("rate limiter wait failed")
			}
			return failedTasks
		}

		log.Trace().Int("idx", task.Idx).Str("url", task.Url).Msg("processing task")

		data, err := pl.Publication.Download(ctx, task.Url)
		if err != nil {
			span.RecordError(err, trace.WithAttributes(attribute.String("url", task.Url)))
			if pl.isCancelled() {
				return failedTasks
			}

			if isRetry {
				log.Error().Err(err).Int("idx", task.Idx).Str("url", task.Url).
					Msg("retry download failed, ending content download")

				select {
				case pl.ErrCh <- fmt.Errorf("final download failed on url %s; %w", task.Url, err):
					pl.Cancel()
				default:
				}
				return failedTasks
			}

			failedTasks = append(failedTasks, task)
			atomic.AddInt64(&pl.Publication.failedDownloads, 1)

			log.Warn().Err(err).Int("idx", task.Idx).Str("url", task.Url).
				Msg("download has failed for the first time, retrying at the end")
			continue
		}

		if pl.isCancelled() {
			return failedTasks
		}

		pl.Publication.speedTracker.IncrementIntermediate()

		select {
		case pl.Publication.iOWorkCh <- ioTask{data, pl.Publication.ContentPath(pl.Chapter), task}:
		case <-pl.Ctx.Done():
			return failedTasks
		}
	}

	return failedTasks
}

type downloadTask struct {
	Idx int
	Url string
}

type ioTask struct {
	Data []byte
	Path string
	Task downloadTask
}

func (p *publication) getChapterById(id string) (Chapter, bool) {
	idx := slices.IndexFunc(p.series.Chapters, func(chapter Chapter) bool {
		return chapter.Id == id
	})

	if idx == -1 {
		p.log.Warn().Str("chapter", id).Int("idx", idx).Msg("chapter not found")
		return Chapter{}, false
	}

	return p.series.Chapters[idx], true
}

func (p *publication) filterContentByUserSelection() {
	if len(p.toDownloadUserSelected) == 0 {
		// Update chapters in case p.Series was updated since

		return
	}

	curSize := len(p.toDownload)
	p.toDownload = utils.MaybeMap(p.series.Chapters, func(chapter Chapter) (string, bool) {
		return chapter.Id, slices.Contains(p.toDownloadUserSelected, chapter.Id)
	})

	p.log.Debug().Int("size", curSize).Int("newSize", len(p.toDownload)).
		Msg("content further filtered after user has made a selection in the UI")

	if len(p.toRemoveContent) == 0 {
		return
	}

	paths := utils.MaybeMap(p.toDownload, func(id string) (string, bool) {
		chapter, ok := p.getChapterById(id)
		if !ok {
			return "", false
		}

		return p.ContentPath(chapter) + ".cbz", true
	})
	p.toRemoveContent = utils.Filter(p.toRemoveContent, func(path string) bool {
		return slices.Contains(paths, path)
	})
}

func (p *publication) startDownloadPipeline(ctx context.Context) {
	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeDownloadContent)
	defer span.End()

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	if hook, ok := p.repository.(PreDownloadHook); ok {
		if err := hook.PreDownloadHook(p, ctx); err != nil {
			p.abortDownload(ctx, err)
			return
		}
	}

	p.filterContentByUserSelection()

	p.log.Info().
		Int("all", len(p.series.Chapters)).
		Int("toDownload", len(p.toDownload)).
		Int("reDownloads", len(p.toRemoveContent)).
		Str("into", p.GetDownloadDir()).
		Msg("downloading content")

	p.speedTracker = utils.NewSpeedTracker(len(p.toDownload))

	start := time.Now()

	p.wg = &sync.WaitGroup{}
	p.ioWg = &sync.WaitGroup{}
	p.iOWorkCh = make(chan ioTask, p.maxImages*2) // Allow for some buffer as I/O may be slower than downloading (webp)

	go p.signalRUpdateLoop(ctx)
	p.StartIOWorkers(ctx)

	if err := p.processDownloads(ctx, p.wg); err != nil {
		p.log.Trace().Err(err).Msg("download failed")
		close(p.iOWorkCh)
		p.abortDownload(ctx, err)
		return
	}

	p.wg.Wait()
	p.log.Debug().Dur("elapsed", time.Since(start)).
		Msg("All content has been downloaded, waiting for I/O workers to finish")

	close(p.iOWorkCh)
	p.SetState(payload.ContentStateCleanup)
	p.ioWg.Wait()

	// Prevent waiting on them again
	p.wg = nil
	p.ioWg = nil

	p.log.Info().Dur("elapsed", time.Since(start)).Msg("Finished downloading content")
	p.StopDownload()
}

// processDownloads loops through all chapters to download, and downloads them. Returning early when one fails
func (p *publication) processDownloads(ctx context.Context, wg *sync.WaitGroup) error {
	for _, chapterId := range p.toDownload {
		chapter, ok := p.getChapterById(chapterId)
		if !ok {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			wg.Add(1)
			err := func() error {
				defer wg.Done()
				return p.downloadChapter(ctx, chapter)
			}()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *publication) downloadChapter(ctx context.Context, chapter Chapter) error {
	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeChapter)
	defer span.End()

	pl, err := p.newPipeline(ctx, chapter)
	if err != nil || pl == nil {
		return err
	}

	defer pl.Cancel()

	chapterPath := p.ContentPath(chapter)
	if err = p.fs.MkdirAll(chapterPath, 0755); err != nil {
		return err
	}

	// Mark as downloaded as soon as the directory is created as we need to remove it in case of an error
	p.hasDownloaded = append(p.hasDownloaded, chapterPath)

	if err = p.writeMetadata(ctx, chapter); err != nil {
		p.log.Warn().Err(err).Msg("failed to write metadata")
	}

	span.AddEvent("write.metadata")
	span.SetAttributes(attribute.Int("size", len(pl.Urls)))

	pl.log.Debug().Int("size", len(pl.Urls)).Msg("starting download")
	start := time.Now()

	p.speedTracker.SetIntermediate(len(pl.Urls))
	go pl.ProduceUrls()

	pl.StartDownloadWorkers() //nolint: contextcheck

	pl.DownloadWg.Wait()
	pl.log.Debug().Dur("elapsed", time.Since(start)).Msg("Finished downloading")

	select {
	case err = <-pl.ErrCh:
		return err
	default:
	}

	if len(pl.Urls) < 5 {
		time.Sleep(1 * time.Second)
	}

	// Reset chapter progress and increment series progress
	p.speedTracker.ClearIntermediate()
	p.speedTracker.Increment()
	return nil
}

func (p *publication) newPipeline(ctx context.Context, chapter Chapter) (*pipeline, error) {
	pipelineCtx, pipelineCancel := context.WithCancel(ctx)

	pl := &pipeline{
		Ctx:         pipelineCtx,
		Cancel:      pipelineCancel,
		Publication: p,
		Chapter:     chapter,
		RateLimiter: rate.NewLimiter(rate.Limit(p.maxImages), p.maxImages),
		DownloadWg:  &sync.WaitGroup{},
		DownloadCh:  make(chan downloadTask, p.maxImages),
		ErrCh:       make(chan error, 1),
		log:         p.ChapterLogger(chapter),
	}

	urls, err := p.repository.ChapterUrls(ctx, chapter)
	if err != nil {
		pl.log.Error().Err(err).Msg("failed to fetch chapter urls")
		return nil, err
	}

	if len(urls) == 0 {
		pl.log.Warn().Msg("content has no downloadable urls? Unexpected? Report this!")
		return nil, nil
	}

	pl.Urls = urls
	return pl, nil
}

func (p *publication) ChapterLogger(chapter Chapter) zerolog.Logger {
	builder := p.log.With().
		Str("chapterId", chapter.Id).
		Str("chapter", chapter.Chapter)

	if chapter.Title != "" {
		builder = builder.Str("title", chapter.Title)
	}

	if chapter.Volume != "" {
		builder = builder.Str("volume", chapter.Volume)
	}

	return builder.Logger()
}

// StartIOWorkers starts cap(IOCh) IOWorker threads
func (p *publication) StartIOWorkers(ctx context.Context) {
	for worker := range cap(p.iOWorkCh) {
		go p.IOWorker(ctx, fmt.Sprintf("IOWorker#%d", worker))
	}
}

// IOWorker reads from IOCh; converts to webp and writes to disk
func (p *publication) IOWorker(ctx context.Context, id string) {
	p.ioWg.Add(1)
	defer p.ioWg.Done()

	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeIOWorker,
		trace.WithAttributes(attribute.String("worker.id", id)))
	defer span.End()

	log := p.log.With().Str("IOWorker#", id).Logger()

	for task := range p.iOWorkCh {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := p.ext.ioTaskFunc(p, ctx, log, task)
		if err != nil {
			p.abortDownload(ctx, fmt.Errorf("unable to run task '%v': %w", task.Path, err))
			return
		}
	}
}

func (p *publication) abortDownload(ctx context.Context, reason error) {
	if errors.Is(reason, context.Canceled) {
		return
	}

	p.log.Error().Err(reason).Msg("error while downloading content")

	if p.cancel != nil {
		p.cancel()
	}

	// Wait for I/O and download tasks
	utils.WaitFor(p.wg, time.Minute*2)
	utils.WaitFor(p.ioWg, time.Minute*2)

	req := payload.StopRequest{
		Provider:    p.Provider(),
		Id:          p.Id(),
		DeleteFiles: true,
	}

	//nolint: contextcheck
	p.notificationService.Notify(context.TODO(), models.NewNotification().
		WithTitle("Failed download").
		WithSummary(fmt.Sprintf("%s failed to download", p.Title())).
		WithBody(fmt.Sprintf("Download failed for %s, because %v", p.Title(), reason)).
		WithGroup(models.GroupError).
		WithColour(models.Error).
		Build())

	if err := p.client.RemoveDownload(req); err != nil {
		p.log.Error().Err(err).Msg("unable to remove download")
	}
}

func (p *publication) signalRUpdateLoop(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.signalR.UpdateContentInfo(p.Request().OwnerId, p.GetInfo())
			}
		}
	}()
}
