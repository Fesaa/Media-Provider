package menou

import (
	"github.com/rs/zerolog"
	"net/http"
	"strconv"
	"time"
)

type retryer struct {
	RoundTripper http.RoundTripper

	log zerolog.Logger
}

func (r *retryer) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if r.RoundTripper == nil {
		r.RoundTripper = http.DefaultTransport
	}

	resp, err = r.RoundTripper.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return resp, nil
	}

	retryAfter := resp.Header.Get("X-RateLimit-Retry-After")
	if retryAfter == "" {
		retryAfter = resp.Header.Get("Retry-After")
	}

	var d time.Duration
	if unix, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
		t := time.Unix(unix, 0)
		d = time.Until(t)
	} else {
		d = time.Minute
	}

	r.log.Warn().
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Str("retryAfter", retryAfter).
		Dur("sleeping_for", d).
		Msg("Too many requests, sleeping and trying again")
	time.Sleep(d)
	return r.RoundTripper.RoundTrip(req)
}
