package subsplease

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const URL string = "https://subsplease.org/api/?f=search&tz=Europe/Brussels&s=%s"

var cache = *utils.NewCache[SearchResult](5 * time.Minute)

type SearchOptions struct {
	Query string
}

func (o SearchOptions) toURL() string {
	return fmt.Sprintf(URL, url.QueryEscape(o.Query))
}

func Search(options SearchOptions) (SearchResult, error) {
	u := options.toURL()
	slog.Debug("Search SubsPlease for anime", "url", u)

	if res := cache.Get(u); res != nil {
		slog.Debug("Cache hit", "url", u)
		return *res, nil
	}

	req, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var r SearchResult
	err = json.Unmarshal(data, &r)
	if err != nil {
		// Subsplease sends back an empty array when no torrents are found
		// instead of an empty map...
		var empty []string
		err2 := json.Unmarshal(data, &empty)
		if err2 == nil {
			return SearchResult{}, nil
		}

		return nil, err
	}

	cache.Set(u, r)
	return r, nil
}
