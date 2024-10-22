package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
)

func Search(req payload.SearchRequest) ([]Info, error) {
	log.Trace("searching", "req", fmt.Sprintf("%+v", req))
	data := make([]Info, 0)
	// A page may have several providers, that don't share the same modifiers
	// So we bottle them up, instead of instantly returning an error
	errors := make([]error, 0)
	for _, p := range req.Provider {
		s, ok := providers[p]
		if !ok {
			log.Warn("provider not supported", "provider", p)
			errors = append(errors, fmt.Errorf("provider %q not supported", p))
			continue
		}

		search, err := s.Search(req)
		if err != nil {
			log.Warn("search error", "provider", p, "error", err)
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

func Download(req payload.DownloadRequest) error {
	log.Trace("downloading", "req", fmt.Sprintf("%+v", req))
	s, ok := providers[req.Provider]
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return s.Download(req)
}

func Stop(req payload.StopRequest) error {
	log.Trace("stopping", "req", fmt.Sprintf("%+v", req))
	s, ok := providers[req.Provider]
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return s.Stop(req)
}
