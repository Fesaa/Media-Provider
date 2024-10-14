package yts

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/wisewolf"
	"io"
)

const URL string = "https://yts.mx/api/v2/list_movies.json?query_term=%s&page=%d&sort_by=%s"

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
	req, err := wisewolf.Client.Get(url)
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
	return &r, nil
}
