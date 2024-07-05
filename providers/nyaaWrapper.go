package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	tools "github.com/Fesaa/go-tools"
	"github.com/irevenko/go-nyaa/nyaa"
	"github.com/irevenko/go-nyaa/types"
	"time"
)

var cache = tools.NewCache[[]types.Torrent](5 * time.Minute)

func cacheKey(opts nyaa.SearchOptions) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", opts.Provider, opts.Filter, opts.SortBy, opts.Category, opts.Query)
}

func nyaaSearch(opts nyaa.SearchOptions) ([]types.Torrent, error) {
	key := cacheKey(opts)
	log.Debug(fmt.Sprintf("Searching %s", opts.Provider), "key", key)

	if hit := cache.Get(key); hit != nil {
		log.Trace("Nyaa Cache hit", "key", key)
		return hit.Get(), nil
	}

	search, err := nyaa.Search(opts)
	if err != nil {
		return nil, err
	}

	cache.Set(key, search)
	return search, nil
}
