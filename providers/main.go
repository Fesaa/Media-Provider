package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	utils2 "github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"sync"
)

func New(log zerolog.Logger, container *dig.Container) *ContentProvider {
	p := &ContentProvider{
		lock:      &sync.Mutex{},
		providers: make(map[models.Provider]Provider),
		log:       log.With().Str("handler", "provider").Logger(),
	}

	utils2.Must(container.Invoke(p.registerAll))
	return p
}

func (p *ContentProvider) Search(req payload.SearchRequest) ([]payload.Info, error) {
	p.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("searching")
	p.lock.Lock()
	defer p.lock.Unlock()

	data := make([]payload.Info, 0)
	// A page may have several providers, that don't share the same modifiers
	// So we bottle them up, instead of instantly returning an error
	errors := make([]error, 0)
	for _, prov := range req.Provider {
		s, ok := p.providers[prov]
		if !ok {
			p.log.Warn().Int("provider", int(prov)).Msg("provider not supported")
			errors = append(errors, fmt.Errorf("provider %q not supported", p))
			continue
		}

		search, err := s.Search(req)
		if err != nil {
			p.log.Warn().Int("provider", int(prov)).Err(err).Msg("searching failed")
			errors = append(errors, fmt.Errorf("provider %q: %w", p, err))
			continue
		}

		data = append(data, search...)
	}

	if len(data) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("no results found: %v", errors)
	}

	return data, nil
}

func (p *ContentProvider) Download(req payload.DownloadRequest) error {
	p.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("downloading")
	p.lock.Lock()
	defer p.lock.Unlock()

	s, ok := p.providers[req.Provider]
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return s.Download(req)
}

func (p *ContentProvider) Stop(req payload.StopRequest) error {
	p.log.Trace().Str("req", fmt.Sprintf("%+v", req)).Msg("stopping")
	p.lock.Lock()
	defer p.lock.Unlock()

	s, ok := p.providers[req.Provider]
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return s.Stop(req)
}
