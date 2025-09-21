package common

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
)

var (
	ErrQueueFull = errors.New("queue is full")
)

type QueueItem interface {
	State() payload.ContentState
	Logger() *zerolog.Logger
	DownloadContent(context.Context)
	LoadMetadata(context.Context)
}

type ProviderQueue interface {
	AddToLoadingQueue(QueueItem) error
	AddToDownloadQueue(QueueItem) error
	Shutdown()
}

// ProviderQueue represents the queue for each separate provider. The queue has just one internal worker
// This worker will always empty out the loadingQueue, then the downloadQueue.
// Each queue has a max capacity of 100. The worker is started automatically after creation
type providerQueue struct {
	log zerolog.Logger

	providerName  models.Provider
	loadingQueue  chan QueueItem
	downloadQueue chan QueueItem

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewProviderQueue(
	provider models.Provider,
	parentCtx context.Context,
	log zerolog.Logger,
) ProviderQueue {

	ctx, cancel := context.WithCancel(parentCtx)

	pq := &providerQueue{
		providerName:  provider,
		loadingQueue:  make(chan QueueItem, 100),
		downloadQueue: make(chan QueueItem, 100),
		ctx:           ctx,
		cancel:        cancel,
		log:           log.With().Any("provider", provider).Logger(),
	}

	go pq.startWorkers()

	return pq
}

// startWorkers starts the work, also manages the WaitGroup
func (pq *providerQueue) startWorkers() {
	pq.wg.Add(1)
	defer pq.wg.Done()

	pq.worker()
}

// worker processes items with loading priority over downloading
func (pq *providerQueue) worker() {
	pq.log.Debug().Msg("worker started")
	defer pq.log.Debug().Msg("worker stopped")

	for {
		select {
		case <-pq.ctx.Done():
			return
		case content := <-pq.loadingQueue:
			pq.processLoadInfo(content)
			continue
		default:
		}

		select {
		case <-pq.ctx.Done():
			return
		case content := <-pq.loadingQueue: // Safeguard in case of edge cases
			pq.processLoadInfo(content)
		case content := <-pq.downloadQueue:
			content.Logger().Trace().Msg("starting download")
			content.DownloadContent(pq.ctx)
		}
	}
}

// processLoadInfo handles loading information for content, this is a blocking operation
func (pq *providerQueue) processLoadInfo(content QueueItem) {
	if content == nil {
		return
	}

	log := content.Logger()

	log.Debug().Msg("starting load info")

	content.LoadMetadata(pq.ctx)

	if content.State() == payload.ContentStateReady {
		log.Trace().Msg("content is ready after loading info, moving to download queue")
		select {
		case pq.downloadQueue <- content:
		case <-pq.ctx.Done():
			return
		}

		return
	}
}

// AddToLoadingQueue adds content to the loading queue
func (pq *providerQueue) AddToLoadingQueue(content QueueItem) error {
	select {
	case pq.loadingQueue <- content:
		return nil
	case <-pq.ctx.Done():
		return pq.ctx.Err()
	default:
		return ErrQueueFull
	}
}

// AddToDownloadQueue adds content to the download queue (for ready content)
func (pq *providerQueue) AddToDownloadQueue(content QueueItem) error {
	select {
	case pq.downloadQueue <- content:
		return nil
	case <-pq.ctx.Done():
		return pq.ctx.Err()
	default:
		return ErrQueueFull
	}
}

// Shutdown gracefully shuts down the provider queue, has a hard limit of 30s
func (pq *providerQueue) Shutdown() {
	pq.log.Debug().Msg("shutting down provider queue")

	pq.cancel()

	close(pq.loadingQueue)
	close(pq.downloadQueue)

	done := make(chan struct{})
	go func() {
		pq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		pq.log.Debug().Msg("provider queue shutdown complete")
	case <-time.After(time.Second * 30):
		pq.log.Warn().Msg("provider queue shutdown timeout")
	}
}
