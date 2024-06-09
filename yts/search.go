package yts

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Fesaa/Media-Provider/utils"
)

const URL string = "https://yts.mx/api/v2/list_movies.json?query_term=%s&page=%d&sort_by=%s"

var cache utils.Cache[YTSSearchResult] = *utils.NewCache[YTSSearchResult](5 * time.Minute)

type YTSSearchOptions struct {
	Query  string
	SortBy string
	Page   int
}

func (o YTSSearchOptions) toURL() string {
	if o.Page == 0 {
		o.Page = 1
	}

	if o.SortBy == "" {
		o.SortBy = "title"
	}

	return fmt.Sprintf(URL, o.Query, o.Page, o.SortBy)
}

func Search(options YTSSearchOptions) (*YTSSearchResult, error) {
	url := options.toURL()
	slog.Info("Searing YTS for movies.", "url", url)

	if res := cache.Get(url); res != nil {
		slog.Debug("Cache hit", "url", url)
		return res, nil
	}

	req, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var r YTSSearchResult
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}

	cache.Set(url, r)
	return &r, nil
}
