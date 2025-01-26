package services

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"time"
)

type ContentService interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	Download(payload.DownloadRequest) error
	DownloadSubscription(*models.Subscription) error
	Stop(payload.StopRequest) error
	RegisterProvider(models.Provider, ProviderAdapter)
	DownloadMetadata(models.Provider) (payload.DownloadMetadata, error)
}

type ProviderAdapter interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
	DownloadMetadata() payload.DownloadMetadata
}

type contentService struct {
	providers *utils.SafeMap[models.Provider, ProviderAdapter]
	log       zerolog.Logger
}

func ContentServiceProvider(log zerolog.Logger) ContentService {
	return &contentService{
		providers: utils.NewSafeMap[models.Provider, ProviderAdapter](),
		log:       log.With().Str("handler", "content-service").Logger(),
	}
}

func (s *contentService) DownloadMetadata(provider models.Provider) (payload.DownloadMetadata, error) {
	adapter, ok := s.providers.Get(provider)
	if !ok {
		return payload.DownloadMetadata{}, fmt.Errorf("provider %q not supported", provider)
	}
	return adapter.DownloadMetadata(), nil
}

func (s *contentService) Search(req payload.SearchRequest) ([]payload.Info, error) {
	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("searching")
	start := time.Now()
	defer func(start time.Time) {
		dur := time.Since(start)
		s.log.Trace().Dur("elapsed", dur).Msg("search has completed")
	}(start)

	// A page may have several providers, that don't share the same modifiers
	// So we bottle them up, instead of instantly returning an error
	var results []payload.Info
	var errs []error

	for _, provider := range req.Provider {
		adapter, ok := s.providers.Get(provider)
		if !ok {
			s.log.Warn().Any("provider", provider).Msg("provider not supported")
			errs = append(errs, fmt.Errorf("provider %d not supported", provider))
			continue
		}

		searchStart := time.Now()
		search, err := adapter.Search(req)
		searchDuration := time.Since(searchStart)
		if err != nil {
			s.log.Warn().Any("provider", provider).Err(err).Msg("searching failed")
			errs = append(errs, fmt.Errorf("provider %d: %w", provider, err))
			continue
		}

		s.log.Debug().Dur("elapsed", searchDuration).Str("request", fmt.Sprintf("%+v", req)).Msg("search done")
		if searchDuration > time.Second*1 {
			s.log.Warn().Dur("elapsed", searchDuration).Msg("searching took more than one second")
		}

		results = append(results, search...)
	}

	if len(results) == 0 && len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return results, nil
}

func (s *contentService) DownloadSubscription(sub *models.Subscription) error {
	return s.Download(payload.DownloadRequest{
		Provider:  sub.Provider,
		Id:        sub.ContentId,
		BaseDir:   sub.Info.BaseDir,
		TempTitle: sub.Info.Title,
	})
}

func (s *contentService) Download(req payload.DownloadRequest) error {
	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("downloading")

	adapter, ok := s.providers.Get(req.Provider)
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}

	return adapter.Download(req)
}

func (s *contentService) Stop(req payload.StopRequest) error {
	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("stopping")

	adapter, ok := s.providers.Get(req.Provider)
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return adapter.Stop(req)
}

func (s *contentService) RegisterProvider(model models.Provider, adapter ProviderAdapter) {
	s.providers.Set(model, adapter)
}
