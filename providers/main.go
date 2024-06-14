package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"log/slog"
)

func Search(req SearchRequest) ([]Info, error) {
	slog.Debug("Searching...", "req", fmt.Sprintf("%+v", req))
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

func Download(req DownloadRequest) error {
	slog.Debug("Downloading...", "req", fmt.Sprintf("%+v", req))
	s, ok := providers[req.Provider]
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return s.Download(req)
}

func Stop(req StopRequest) error {
	slog.Debug("Stopping...", "req", fmt.Sprintf("%+v", req))
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
