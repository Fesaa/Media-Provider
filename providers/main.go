package providers

import "fmt"

func Search(req SearchRequest) ([]TorrentInfo, error) {
	s, ok := providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("provider %q not supported", req.Provider)
	}

	return s.Search(req)
}
