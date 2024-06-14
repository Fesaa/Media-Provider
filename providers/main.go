package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
)

func Search(req SearchRequest) ([]TorrentInfo, error) {
	s, ok := providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("provider %q not supported", req.Provider)
	}

	return s.Search(req)
}

func Download(req DownloadRequest) error {
	s, ok := providers[req.Provider]
	if !ok {
		return fmt.Errorf("provider %q not supported", req.Provider)
	}
	return s.Download(req)
}

func HasProvider(provider config.Provider) bool {
	_, ok := providers[provider]
	return ok
}
