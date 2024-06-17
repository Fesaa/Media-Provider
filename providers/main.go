package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
)

func Search(req payload.SearchRequest) ([]Info, error) {
	log.Trace("searching", "req", fmt.Sprintf("%+v", req))
	data := make([]Info, 0)
	for _, p := range req.Provider {
		s, ok := providers[p]
		if !ok {
			return nil, fmt.Errorf("provider %q not supported", req.Provider)
		}

		search, err := s.Search(req)
		if err != nil {
			return nil, err
		}

		data = append(data, search...)
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

func HasProvider(provider config.Provider) bool {
	_, ok := providers[provider]
	return ok
}
