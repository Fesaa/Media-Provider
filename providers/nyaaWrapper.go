package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/irevenko/go-nyaa/nyaa"
	"github.com/irevenko/go-nyaa/types"
)

func cacheKey(opts nyaa.SearchOptions) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", opts.Provider, opts.Filter, opts.SortBy, opts.Category, opts.Query)
}

func nyaaSearch(opts nyaa.SearchOptions) ([]types.Torrent, error) {
	key := cacheKey(opts)
	log.Debug(fmt.Sprintf("Searching %s", opts.Provider), "key", key)

	search, err := nyaa.Search(opts)
	if err != nil {
		return nil, err
	}
	return search, nil
}
