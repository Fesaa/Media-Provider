package pasloe

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/utils"
	"go.uber.org/dig"
	"net/http"
	"sync"
)

type Registry interface {
	Create(c api.Client, req payload.DownloadRequest) (api.Downloadable, error)
}

type registry struct {
	r          map[models.Provider]func(scope *dig.Scope) api.Downloadable
	mu         sync.RWMutex
	httpClient *http.Client
	container  *dig.Container
}

func newRegistry(httpClient *http.Client, container *dig.Container) Registry {
	r := &registry{
		r:          make(map[models.Provider]func(scope *dig.Scope) api.Downloadable),
		mu:         sync.RWMutex{},
		httpClient: httpClient,
		container:  container,
	}

	r.Register(models.WEBTOON, webtoon.NewWebToon)
	r.Register(models.MANGADEX, mangadex.NewManga)
	r.Register(models.DYNASTY, dynasty.NewManga)

	return r
}

func (r *registry) Register(provider models.Provider, fn func(scope *dig.Scope) api.Downloadable) {
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

	scope := r.container.Scope("pasloe::registry::create")

	utils.Must(scope.Provide(utils.Identity(c)))
	utils.Must(scope.Provide(utils.Identity(req)))

	return fn(scope), nil
}
