package pasloe

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/bato"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"go.uber.org/dig"
)

type Registry interface {
	Create(c publication.Client, req payload.DownloadRequest) (publication.Publication, error)
}

type registry struct {
	container *dig.Container
}

func newRegistry(container *dig.Container) Registry {
	return &registry{
		container: container,
	}
}

func (r *registry) Create(c publication.Client, req payload.DownloadRequest) (publication.Publication, error) {
	scope := r.container.Scope("pasloe::registry::create")

	utils.Must(scope.Provide(utils.Identity(c)))
	utils.Must(scope.Provide(utils.Identity(req)))
	// TODO: This needs to be updated if we have stuff other than cbz
	utils.Must(scope.Provide(utils.Identity(publication.CbzExt())))

	switch req.Provider { //nolint: exhaustive
	case models.BATO:
		utils.ProviderAs[bato.Repository, publication.Repository](scope, bato.NewRepository)
	case models.DYNASTY:
		utils.ProviderAs[dynasty.Repository, publication.Repository](scope, dynasty.NewRepository)
	case models.MANGADEX:
		utils.ProviderAs[mangadex.Repository, publication.Repository](scope, mangadex.NewRepository)
	case models.WEBTOON:
		utils.ProviderAs[webtoon.Repository, publication.Repository](scope, webtoon.NewRepository)
	default:
		return nil, services.ErrProviderNotSupported
	}

	utils.Must(scope.Provide(publication.New))
	return utils.MayInvoke[publication.Publication](scope)
}
