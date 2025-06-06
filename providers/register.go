package providers

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/bato"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/limetorrents"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/nyaa"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/subsplease"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/yts"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

func RegisterProviders(s services.ContentService, container *dig.Container) {
	scope := container.Scope("content-providers")

	utils.Must(container.Provide(webtoon.NewRepository))
	utils.Must(container.Provide(mangadex.NewRepository))
	utils.Must(container.Provide(dynasty.NewRepository))
	utils.Must(container.Provide(bato.NewRepository))

	utils.Must(scope.Provide(yts.NewBuilder))
	utils.Must(scope.Provide(subsplease.NewBuilder))
	utils.Must(scope.Provide(limetorrents.NewBuilder))
	utils.Must(scope.Provide(webtoon.NewBuilder))
	utils.Must(scope.Provide(mangadex.NewBuilder))
	utils.Must(scope.Provide(dynasty.NewBuilder))
	utils.Must(scope.Provide(nyaa.NewBuilder))
	utils.Must(scope.Provide(bato.NewBuilder))

	utils.Must(registerProviderAdapter[*yts.Builder](s, scope))
	utils.Must(registerProviderAdapter[*subsplease.Builder](s, scope))
	utils.Must(registerProviderAdapter[*limetorrents.Builder](s, scope))
	utils.Must(registerProviderAdapter[*webtoon.Builder](s, scope))
	utils.Must(registerProviderAdapter[*mangadex.Builder](s, scope))
	utils.Must(registerProviderAdapter[*dynasty.Builder](s, scope))
	utils.Must(registerProviderAdapter[*nyaa.Builder](s, scope))
	utils.Must(registerProviderAdapter[*bato.Builder](s, scope))
}

type defaultProviderAdapter[T any, S any] struct {
	transformer func(payload.SearchRequest) S
	searcher    func(S) (T, error)
	normalizer  func(T) []payload.Info
	metadata    func() payload.DownloadMetadata
	client      func() services.Client
	provider    models.Provider
	log         zerolog.Logger
}

func registerProviderAdapter[B builder[T, S], T, S any](s services.ContentService, scope *dig.Scope) error {
	return scope.Invoke(func(builder B) {
		reqMapper := &defaultProviderAdapter[T, S]{
			transformer: builder.Transform,
			normalizer:  builder.Normalize,
			searcher:    builder.Search,
			provider:    builder.Provider(),
			metadata:    builder.DownloadMetadata,
			log:         builder.Logger(),
			client:      builder.Client,
		}

		s.RegisterProvider(builder.Provider(), reqMapper)
	})
}

type builder[T, S any] interface {
	Provider() models.Provider
	Logger() zerolog.Logger
	Normalize(T) []payload.Info
	Transform(payload.SearchRequest) S
	Search(S) (T, error)
	DownloadMetadata() payload.DownloadMetadata
	Client() services.Client
}

func (s *defaultProviderAdapter[T, S]) Search(req payload.SearchRequest) ([]payload.Info, error) {
	t := s.transformer(req)
	data, err := s.searcher(t)
	if err != nil {
		return nil, err
	}
	return s.normalizer(data), nil
}

func (s *defaultProviderAdapter[T, S]) DownloadMetadata() payload.DownloadMetadata {
	return s.metadata()
}

func (s *defaultProviderAdapter[T, S]) Client() services.Client {
	return s.client()
}
