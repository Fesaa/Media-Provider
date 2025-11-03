package pasloe

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/bato"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangabuddy"
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

	err := utils.Errs(
		scope.Provide(publication.New),
		scope.Provide(utils.Identity(c)),
		scope.Provide(utils.Identity(req)),
		// TODO: This needs to be updated if we have stuff other than cbz
		scope.Provide(utils.Identity(publication.CbzExt())),
	)

	if err != nil {
		return nil, err
	}

	switch req.Provider { //nolint: exhaustive
	case models.BATO:
		err = utils.ProviderAs[bato.Repository, publication.Repository](scope, bato.NewRepository)
	case models.DYNASTY:
		err = utils.ProviderAs[dynasty.Repository, publication.Repository](scope, dynasty.NewRepository)
	case models.MANGADEX:
		err = utils.ProviderAs[mangadex.Repository, publication.Repository](scope, mangadex.NewRepository)
	case models.WEBTOON:
		err = utils.ProviderAs[webtoon.Repository, publication.Repository](scope, webtoon.NewRepository)
	case models.MANGA_BUDDY:
		err = utils.ProviderAs[mangabuddy.Repository, publication.Repository](scope, mangabuddy.NewRepository)
	default:
		return nil, services.ErrProviderNotSupported
	}

	if err != nil {
		return nil, err
	}

	return utils.MayInvoke[publication.Publication](scope)
}
