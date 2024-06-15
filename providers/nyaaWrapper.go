package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/irevenko/go-nyaa/nyaa"
	"github.com/irevenko/go-nyaa/types"
	"log/slog"
	"time"
)

var cache = *utils.NewCache[[]types.Torrent](5 * time.Minute)

func cacheKey(opts nyaa.SearchOptions) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s", opts.Provider, opts.Filter, opts.SortBy, opts.Category, opts.Query)
}

func nyaaSearch(opts nyaa.SearchOptions) ([]types.Torrent, error) {
	key := cacheKey(opts)
	slog.Debug(fmt.Sprintf("Searching %s", opts.Provider), "key", key)

	if hit := cache.Get(key); hit != nil {
		return *hit, nil
	}

	search, err := nyaa.Search(opts)
	if err != nil {
		return nil, err
	}

	cache.Set(key, search)
	return search, nil
}
