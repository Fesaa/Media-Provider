package subsplease

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

const URL string = "https://subsplease.org/api/?f=search&tz=Europe/Brussels&s=%s"

type SearchOptions struct {
	Query string
}

func (o SearchOptions) toURL() string {
	return fmt.Sprintf(URL, url.QueryEscape(o.Query))
}

func (b *Builder) Search(options SearchOptions) (SearchResult, error) {
	u := options.toURL()
	req, err := b.httpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			b.log.Warn().Err(err).Msg("failed to close body")
		}
	}(req.Body)

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

	return r, nil
}
