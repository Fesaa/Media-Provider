package providers

import (
	"github.com/irevenko/go-nyaa/nyaa"
	"github.com/irevenko/go-nyaa/types"
)

func nyaaSearch(opts nyaa.SearchOptions) ([]types.Torrent, error) {

	search, err := nyaa.Search(opts)
	if err != nil {
		return nil, err
	}
	return search, nil
}
