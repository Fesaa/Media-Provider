package pasloe

import (
	"errors"
	"sync"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/bato"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/utils"
	"go.uber.org/dig"
)

type Registry interface {
	Create(c core.Client, req payload.DownloadRequest) (core.Downloadable, error)
}

type registry struct {
	r         map[models.Provider]func(scope *dig.Scope) core.Downloadable
	mu        sync.RWMutex
	container *dig.Container
}

func newRegistry(container *dig.Container) Registry {
	r := &registry{
		r:         make(map[models.Provider]func(scope *dig.Scope) core.Downloadable),
		mu:        sync.RWMutex{},
		container: container,
	}

	r.Register(models.WEBTOON, webtoon.New)
	r.Register(models.MANGADEX, mangadex.New)
	r.Register(models.DYNASTY, dynasty.New)
	r.Register(models.BATO, bato.New)

	return r
}

func (r *registry) Register(provider models.Provider, fn func(scope *dig.Scope) core.Downloadable) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.r[provider] = fn
}

func (r *registry) Create(c core.Client, req payload.DownloadRequest) (core.Downloadable, error) {
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
