package pasloe

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"sync"
)

type registry struct {
	r  map[models.Provider]func(payload.DownloadRequest, api.Client) api.Downloadable
	mu sync.RWMutex
}

func newRegistry() *registry {
	r := &registry{
		r:  make(map[models.Provider]func(payload.DownloadRequest, api.Client) api.Downloadable),
		mu: sync.RWMutex{},
	}

	r.Register(models.WEBTOON, webtoon.NewWebToon)
	r.Register(models.MANGADEX, mangadex.NewManga)

	return r
}

func (r *registry) Register(provider models.Provider, fn func(payload.DownloadRequest, api.Client) api.Downloadable) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.r[provider] = fn
}

func (r *registry) Create(c api.Client, req payload.DownloadRequest) (api.Downloadable, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	fn, ok := r.r[req.Provider]
	if !ok {
		return nil, errors.New("unknown provider")
	}
	return fn(req, c), nil
}
