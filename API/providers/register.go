package providers

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/providers/pasloe/bato"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangabuddy"
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

func RegisterProviders(s services.ContentService, container *dig.Container) error {
	scope := container.Scope("content-providers")

	return utils.Errs(
		container.Provide(webtoon.NewRepository),
		container.Provide(mangadex.NewRepository),
		container.Provide(dynasty.NewRepository),
		container.Provide(bato.NewRepository),
		container.Provide(mangabuddy.NewRepository),

		scope.Provide(yts.NewBuilder),
		scope.Provide(subsplease.NewBuilder),
		scope.Provide(limetorrents.NewBuilder),
		scope.Provide(webtoon.NewBuilder),
		scope.Provide(mangadex.NewBuilder),
		scope.Provide(dynasty.NewBuilder),
		scope.Provide(nyaa.NewBuilder),
		scope.Provide(bato.NewBuilder),
		scope.Provide(mangabuddy.NewBuilder),

		registerProviderAdapter[*yts.Builder](s, scope),
		registerProviderAdapter[*subsplease.Builder](s, scope),
		registerProviderAdapter[*limetorrents.Builder](s, scope),
		registerProviderAdapter[*webtoon.Builder](s, scope),
		registerProviderAdapter[*mangadex.Builder](s, scope),
		registerProviderAdapter[*dynasty.Builder](s, scope),
		registerProviderAdapter[*nyaa.Builder](s, scope),
		registerProviderAdapter[*bato.Builder](s, scope),
		registerProviderAdapter[*mangabuddy.Builder](s, scope),
	)
}

type defaultProviderAdapter[T any, S any] struct {
	transformer func(context.Context, payload.SearchRequest) S
	searcher    func(context.Context, S) (T, error)
	normalizer  func(context.Context, T) []payload.Info
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
	Normalize(context.Context, T) []payload.Info
	Transform(context.Context, payload.SearchRequest) S
	Search(context.Context, S) (T, error)
	DownloadMetadata() payload.DownloadMetadata
	Client() services.Client
}

func (s *defaultProviderAdapter[T, S]) Search(ctx context.Context, req payload.SearchRequest) ([]payload.Info, error) {
	transformCtx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesContentSearch+".transform")
	t := s.transformer(transformCtx, req)
	span.End()

	searchCtx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesContentSearch+".search")
	data, err := s.searcher(searchCtx, t)
	if err != nil {
		span.RecordError(err)
		span.End()
		return nil, err
	}
	span.End()

	normalizeCtx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesContentSearch+".normalize")
	defer span.End()

	return s.normalizer(normalizeCtx, data), nil
}

func (s *defaultProviderAdapter[T, S]) DownloadMetadata() payload.DownloadMetadata {
	return s.metadata()
}

func (s *defaultProviderAdapter[T, S]) Client() services.Client {
	return s.client()
}
