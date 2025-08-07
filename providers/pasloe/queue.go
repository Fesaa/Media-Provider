package pasloe

import (
	"context"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
	"sync"
	"time"
)

type ProviderQueue struct {
	providerName  models.Provider
	loadingQueue  chan core.Downloadable
	downloadQueue chan core.Downloadable
	ctx           context.Context
	cancel        context.CancelFunc
	log           zerolog.Logger
	client        *client
	wg            sync.WaitGroup
}

func NewProviderQueue(provider models.Provider, parentCtx context.Context, client *client) *ProviderQueue {
	ctx, cancel := context.WithCancel(parentCtx)

	pq := &ProviderQueue{
		providerName:  provider,
		loadingQueue:  make(chan core.Downloadable, 100),
		downloadQueue: make(chan core.Downloadable, 100),
		ctx:           ctx,
		cancel:        cancel,
		log:           client.log.With().Any("provider", provider).Logger(),
		client:        client,
	}

	go pq.startWorkers()

	return pq
}

// startWorkers starts the loading and download workers
func (pq *ProviderQueue) startWorkers() {
	pq.wg.Add(1)
	defer pq.wg.Done()

	pq.worker()
}

// worker processes items with loading priority over downloading
func (pq *ProviderQueue) worker() {
	pq.log.Debug().Msg("priority worker started")
	defer pq.log.Debug().Msg("priority worker stopped")

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
			content.Logger().Debug().Msg("starting download")
			content.DownloadContent(pq.ctx)
		}
	}
}

// processLoadInfo handles loading information for content
func (pq *ProviderQueue) processLoadInfo(content core.Downloadable) {
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
func (pq *ProviderQueue) AddToLoadingQueue(content core.Downloadable) error {
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
func (pq *ProviderQueue) AddToDownloadQueue(content core.Downloadable) error {
	select {
	case pq.downloadQueue <- content:
		return nil
	case <-pq.ctx.Done():
		return pq.ctx.Err()
	default:
		return services.ErrQueueFull
	}
}

// Shutdown gracefully shuts down the provider queue
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
