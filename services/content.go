package services

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/dynasty"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/pasloe/nyaa"
	"github.com/Fesaa/Media-Provider/providers/pasloe/webtoon"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/limetorrents"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/subsplease"
	"github.com/Fesaa/Media-Provider/providers/yoitsu/yts"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"time"
)

type ContentService interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}

type requestMapper interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}

type builder[T, S any] interface {
	Provider() models.Provider
	Logger() zerolog.Logger
	Normalize(T) []payload.Info
	Transform(payload.SearchRequest) S
	Search(S) (T, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}

type contentService struct {
	providers *utils.SafeMap[models.Provider, requestMapper]
	log       zerolog.Logger
}

func ContentServiceProvider(container *dig.Container, log zerolog.Logger) ContentService {
	service := &contentService{
		providers: utils.NewSafeMap[models.Provider, requestMapper](),
		log:       log,
	}

	service.registerAll(container)
	return service
}

func (s *contentService) Search(req payload.SearchRequest) ([]payload.Info, error) {
	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("searching")

	data := make([]payload.Info, 0)
	// A page may have several providers, that don't share the same modifiers
	// So we bottle them up, instead of instantly returning an error
	errors := make([]error, 0)

	for _, provider := range req.Provider {
		reqMapper, ok := s.providers.Get(provider)
		if !ok {
			s.log.Warn().Int("provider", int(provider)).Msg("provider not supported")
			errors = append(errors, fmt.Errorf("provider %d not supported", provider))
			continue
		}

		search, err := reqMapper.Search(req)
		if err != nil {
			s.log.Warn().Int("provider", int(provider)).Err(err).Msg("searching failed")
			errors = append(errors, fmt.Errorf("provider %d: %w", provider, err))
			continue
		}

		data = append(data, search...)
	}

	if len(data) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("no results found: %v", errors)
	}

	return data, nil
}

func (s *contentService) Download(req payload.DownloadRequest) error {
	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("downloading")

	reqMapper, ok := s.providers.Get(req.Provider)
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}

	return reqMapper.Download(req)
}

func (s *contentService) Stop(req payload.StopRequest) error {
	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("stopping")

	reqMapper, ok := s.providers.Get(req.Provider)
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return reqMapper.Stop(req)
}

func (s *contentService) registerAll(container *dig.Container) {
	scope := container.Scope("content-providers")

	utils.Must(scope.Provide(yts.NewBuilder))
	utils.Must(scope.Provide(subsplease.NewBuilder))
	utils.Must(scope.Provide(limetorrents.NewBuilder))
	utils.Must(scope.Provide(webtoon.NewBuilder))
	utils.Must(scope.Provide(mangadex.NewBuilder))
	utils.Must(scope.Provide(dynasty.NewBuilder))
	utils.Must(scope.Provide(nyaa.NewNyaaBuilder))

	utils.Must(registerRequestMapper[*yts.Builder](s, scope))
	utils.Must(registerRequestMapper[*subsplease.Builder](s, scope))
	utils.Must(registerRequestMapper[*limetorrents.Builder](s, scope))
	utils.Must(registerRequestMapper[*webtoon.Builder](s, scope))
	utils.Must(registerRequestMapper[*mangadex.Builder](s, scope))
	utils.Must(registerRequestMapper[*dynasty.Builder](s, scope))
	utils.Must(registerRequestMapper[*nyaa.Builder](s, scope))
}

func registerRequestMapper[B builder[T, S], T, S any](s *contentService, scope *dig.Scope) error {
	return scope.Invoke(func(builder B) {

		reqMapper := &requestMapperImpl[T, S]{
			transformer: builder.Transform,
			normalizer:  builder.Normalize,
			searcher:    builder.Search,
			downloader:  builder.Download,
			stopper:     builder.Stop,
			provider:    builder.Provider(),
			log:         builder.Logger(),
		}

		s.providers.Set(builder.Provider(), reqMapper)
	})
}

type requestMapperImpl[T any, S any] struct {
	transformer func(payload.SearchRequest) S
	searcher    func(S) (T, error)
	normalizer  func(T) []payload.Info
	downloader  func(payload.DownloadRequest) error
	stopper     func(payload.StopRequest) error
	provider    models.Provider
	log         zerolog.Logger
}

func (s *requestMapperImpl[T, S]) Download(req payload.DownloadRequest) error {
	return s.downloader(req)
}

func (s *requestMapperImpl[T, S]) Stop(req payload.StopRequest) error {
	return s.stopper(req)
}

func (s *requestMapperImpl[T, S]) Search(req payload.SearchRequest) ([]payload.Info, error) {
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
