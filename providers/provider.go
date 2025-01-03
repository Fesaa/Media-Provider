package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/limetorrents"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/subsplease"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/yts"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"time"
)

func (p *ContentProvider) registerAll(container *dig.Container) {
	scope := container.Scope("content-providers")

	utils.Must(scope.Provide(yts.NewBuilder))
	utils.Must(scope.Provide(subsplease.NewBuilder))
	utils.Must(scope.Provide(limetorrents.NewBuilder))
	utils.Must(scope.Provide(webtoon.NewBuilder))
	utils.Must(scope.Provide(mangadex.NewBuilder))
	utils.Must(scope.Provide(dynasty.NewBuilder))
	utils.Must(scope.Provide(NewNyaaBuilder))

	utils.Must(scope.Invoke(func(builder *yts.Builder) {
		p.register(NewProvider(builder))
	}))
	utils.Must(scope.Invoke(func(builder *subsplease.Builder) {
		p.register(NewProvider(builder))
	}))
	utils.Must(scope.Invoke(func(builder *limetorrents.Builder) {
		p.register(NewProvider(builder))
	}))
	utils.Must(scope.Invoke(func(builder *webtoon.Builder) {
		p.register(NewProvider(builder))
	}))
	utils.Must(scope.Invoke(func(builder *mangadex.Builder) {
		p.register(NewProvider(builder))
	}))
	utils.Must(scope.Invoke(func(builder *dynasty.Builder) {
		p.register(NewProvider(builder))
	}))
	utils.Must(scope.Invoke(func(builder *NyaaBuilder) {
		p.register(NewProvider(builder))
	}))
}

func (p *ContentProvider) register(name models.Provider, provider Provider) {
	p.providers[name] = provider
}

func NewProvider[T, S any](builder ProviderBuilder[T, S]) (models.Provider, Provider) {
	return builder.Provider(), &providerImpl[T, S]{
		transformer: builder.Transform,
		normalizer:  builder.Normalize,
		searcher:    builder.Search,
		downloader:  builder.Download,
		stopper:     builder.Stop,
		provider:    builder.Provider(),
		log:         builder.Logger(),
	}
}

type responseNormalizerFunc[T any] func(T) []payload.Info
type requestTransformerFunc[S any] func(payload.SearchRequest) S
type searchFunc[S, T any] func(S) (T, error)
type downloadFunc func(payload.DownloadRequest) error
type stopFunc func(payload.StopRequest) error

type providerImpl[T any, S any] struct {
	transformer requestTransformerFunc[S]
	normalizer  responseNormalizerFunc[T]
	searcher    searchFunc[S, T]
	downloader  downloadFunc
	stopper     stopFunc
	provider    models.Provider
	log         zerolog.Logger
}

func (s *providerImpl[T, S]) Download(req payload.DownloadRequest) error {
	return s.downloader(req)
}

func (s *providerImpl[T, S]) Stop(req payload.StopRequest) error {
	return s.stopper(req)
}

func (s *providerImpl[T, S]) Search(req payload.SearchRequest) ([]payload.Info, error) {
	t := s.transformer(req)

	start := time.Now()
	data, err := s.searcher(t)
	since := time.Since(start)

	s.log.Debug().Dur("elapsed", since).Str("request", fmt.Sprintf("%+v", req)).Msg("search done")
	if since > time.Second*1 {
		s.log.Warn().Dur("elapsed", since).Msg("searching took more than one second")
	}

	if err != nil {
		s.log.Error().Err(err).Msg("search failed")
		return nil, err
	}
	return s.normalizer(data), nil
}
