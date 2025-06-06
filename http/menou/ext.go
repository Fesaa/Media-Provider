package menou

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
)

func (c *Client) WrapInDoc(ctx context.Context, url string, f ...func(*http.Request) error) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(f) > 0 {
		if err := f[0](req); err != nil {
			return nil, err
		}
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			c.log.Warn().Err(err).Msg("failed to close body")
		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
