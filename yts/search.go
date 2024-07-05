package yts

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	tools "github.com/Fesaa/go-tools"
	"io"
	"net/http"
	"time"
)

const URL string = "https://yts.mx/api/v2/list_movies.json?query_term=%s&page=%d&sort_by=%s"

var cache = tools.NewCache[*SearchResult](5 * time.Minute)

type SearchOptions struct {
	Query  string
	SortBy string
	Page   int
}

func (o SearchOptions) toURL() string {
	if o.Page == 0 {
		o.Page = 1
	}

	if o.SortBy == "" {
		o.SortBy = "title"
	}

	return fmt.Sprintf(URL, o.Query, o.Page, o.SortBy)
}

func Search(options SearchOptions) (*SearchResult, error) {
	url := options.toURL()
	log.Debug("Searing YTS for movies", "url", url)

	if res := cache.Get(url); res != nil {
		log.Trace("YTS cache hit", "url", url)
		return res.Get(), nil
	}

	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Warn("failed to close body", "error", err)
		}
	}(req.Body)

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var r SearchResult
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	cache.Set(url, &r)
	return &r, nil
}
