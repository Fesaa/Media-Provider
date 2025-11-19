package yts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	BaseUrl = "https://yts.lt"
	URL     = BaseUrl + "/api/v2/list_movies.json?query_term=%s&page=%d&sort_by=%s"
)

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

	return fmt.Sprintf(URL, url.QueryEscape(o.Query), o.Page, o.SortBy)
}

func (b *Builder) Search(ctx context.Context, options SearchOptions) (*SearchResult, error) {
	uri := options.toURL()
	req, err := b.httpClient.Get(uri)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			b.log.Warn().Err(err).Msg("failed to close response body")
		}
	}(req.Body)

	if req.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status code %d", req.StatusCode)
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var r SearchResult
	err = json.Unmarshal(data, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
