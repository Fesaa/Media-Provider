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

var (
	ErrProviderNotSupported = errors.New("provider not supported")
	ErrContentAlreadyExists = errors.New("content already exists")
	ErrContentNotFound      = errors.New("content not found")
)

type ContentService interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	Download(payload.DownloadRequest) error
	DownloadSubscription(*models.Subscription) error
	Stop(payload.StopRequest) error
	RegisterProvider(models.Provider, ProviderAdapter)
	DownloadMetadata(models.Provider) (payload.DownloadMetadata, error)
	Message(payload.Message) (payload.Message, error)
}

type Content interface {
	// Id returns a string that uniquely identifies this Content
	Id() string
	// Title returns the title of the Content, depending on how much info has been loaded, this may be equal to Id
	Title() string
	// Provider returns the provider this Content uses
	Provider() models.Provider
	// GetInfo returns the information to be displayed in the UI
	GetInfo() payload.InfoStat
	// State returns the payload.ContentState this Content currently is in
	State() payload.ContentState
	// Message passes any payload.Message to the Content, and returns its response
	// The logic of what happens with the message may depend on the underlying Content
	Message(payload.Message) (payload.Message, error)
}

type Client interface {
	// Download starts the download of the Content associated with the id. The returned error
	// might be because of the Content already have been added, or because of an error with the data
	// provider to create the Content
	Download(payload.DownloadRequest) error
	// RemoveDownload stops the download of the Content associated with the id. Will start cleanup,
	// should only return an error if the id has no associated Content
	RemoveDownload(payload.StopRequest) error
	// Content returns the Content associated with the passed id. If none is found, returns nil
	Content(string) Content
}

type ProviderAdapter interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	DownloadMetadata() payload.DownloadMetadata
	Client() Client
}

type contentService struct {
	providers utils.SafeMap[models.Provider, ProviderAdapter]
	log       zerolog.Logger
}

func ContentServiceProvider(log zerolog.Logger) ContentService {
	return &contentService{
		providers: utils.NewSafeMap[models.Provider, ProviderAdapter](),
		log:       log.With().Str("handler", "content-service").Logger(),
	}
}

func (s *contentService) Message(message payload.Message) (payload.Message, error) {
	adapter, ok := s.providers.Get(message.Provider)
	if !ok {
		return payload.Message{}, ErrProviderNotSupported
	}

	content := adapter.Client().Content(message.ContentId)
	if content == nil {
		return payload.Message{}, ErrContentNotFound
	}

	return content.Message(message)
}

func (s *contentService) DownloadMetadata(provider models.Provider) (payload.DownloadMetadata, error) {
	adapter, ok := s.providers.Get(provider)
	if !ok {
		return payload.DownloadMetadata{}, ErrProviderNotSupported
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
			errs = append(errs, fmt.Errorf("searched failed for provider %d: %w", provider, err))
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

	if len(errs) > 0 {
		s.log.Warn().Err(errors.Join(errs...)).
			Msg("At least one error occurred while searching, retuning found results")
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
	adapter, ok := s.providers.Get(req.Provider)
	if !ok {
		return ErrProviderNotSupported
	}

	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("downloading")
	return adapter.Client().Download(req)
}

func (s *contentService) Stop(req payload.StopRequest) error {
	adapter, ok := s.providers.Get(req.Provider)
	if !ok {
		return ErrProviderNotSupported
	}

	s.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("stopping")
	return adapter.Client().RemoveDownload(req)
}

func (s *contentService) RegisterProvider(provider models.Provider, adapter ProviderAdapter) {
	if s.providers.Has(provider) {
		s.log.Warn().Any("provider", provider).Msg("provider already registered")
	}

	s.providers.Set(provider, adapter)
}
