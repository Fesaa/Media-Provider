package pasloe

import (
	"context"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
)

// ProviderQueue represents the queue for each separate provider. The queue has just one internal worker
// This worker will always empty out the loadingQueue, then the downloadQueue.
// Each queue has a max capacity of 100. The worker is started automatically after creation
type ProviderQueue struct {
	log    zerolog.Logger
	client publication.Client

	providerName  models.Provider
	loadingQueue  chan publication.Publication
	downloadQueue chan publication.Publication

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewProviderQueue(
	provider models.Provider,
	parentCtx context.Context,
	client publication.Client,
	log zerolog.Logger,
) *ProviderQueue {

	ctx, cancel := context.WithCancel(parentCtx)

	pq := &ProviderQueue{
		providerName:  provider,
		loadingQueue:  make(chan publication.Publication, 100),
		downloadQueue: make(chan publication.Publication, 100),
		ctx:           ctx,
		cancel:        cancel,
		log:           log.With().Any("provider", provider).Logger(),
		client:        client,
	}

	go pq.startWorkers()

	return pq
}

// startWorkers starts the work, also manages the WaitGroup
func (pq *ProviderQueue) startWorkers() {
	pq.wg.Add(1)
	defer pq.wg.Done()

	pq.worker()
}

// worker processes items with loading priority over downloading
func (pq *ProviderQueue) worker() {
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
			content.DownloadContent(pq.ctx)
		}
	}
}

// processLoadInfo handles loading information for content, this is a blocking operation
func (pq *ProviderQueue) processLoadInfo(content publication.Publication) {
	if content == nil {
		return
	}

	log := pq.log.With().
		Str("id", content.Id()).
		Str("title", content.Title()).
		Logger()

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
func (pq *ProviderQueue) AddToLoadingQueue(content publication.Publication) error {
	select {
	case pq.loadingQueue <- content:
		return nil
	case <-pq.ctx.Done():
		return pq.ctx.Err()
	default:
		return services.ErrQueueFull
	}
}

// AddToDownloadQueue adds content to the download queue (for ready content)
func (pq *ProviderQueue) AddToDownloadQueue(content publication.Publication) error {
	select {
	case pq.downloadQueue <- content:
		return nil
	case <-pq.ctx.Done():
		return pq.ctx.Err()
	default:
		return services.ErrQueueFull
	}
}

// Shutdown gracefully shuts down the provider queue, has a hard limit of 30s
func (pq *ProviderQueue) Shutdown() {
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
