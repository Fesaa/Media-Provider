package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (c *Core[C, S]) Download(ctx context.Context, url string, tryAgain ...bool) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if customizer, ok := c.impl.(DownloadCustomizer); ok {
		if err := customizer.CustomizeRequest(req); err != nil {
			return nil, err
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			c.Log.Warn().Err(err).Msg("error closing body")
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	if len(tryAgain) > 0 && !tryAgain[0] {
		return nil, fmt.Errorf("hit rate limit too many times")
	}

	retryAfter := resp.Header.Get("X-RateLimit-Retry-After")
	var d time.Duration

	if unix, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
		t := time.Unix(unix, 0)
		d = time.Until(t)
	} else {
		d = time.Minute
	}

	c.Log.Warn().Dur("sleeping_for", d).Msg("Hit rate limit, sleeping")
	time.Sleep(d)
	return c.Download(ctx, url, false)
}

func (c *Core[C, S]) DownloadAndWrite(ctx context.Context, url string, filePath string, tryAgain ...bool) error {
	data, err := c.Download(ctx, url, tryAgain...)
	if err != nil {
		return err
	}

	if err = c.fs.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	return nil
}
