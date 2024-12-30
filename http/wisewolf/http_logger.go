package wisewolf

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

type loggingTransport struct {
	Transport http.RoundTripper

	log zerolog.Logger
}

func (lt *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if lt.Transport == nil {
		lt.Transport = http.DefaultTransport
	}

	startTime := time.Now()
	resp, err := lt.Transport.RoundTrip(req)
	duration := time.Since(startTime)

	l := lt.log.With().
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Dur("duration", duration).
		Logger()

	if err != nil {
		l.Trace().Err(err).Msg("http request returned a non-nil error")
		return resp, err
	}

	if resp.StatusCode >= 400 {
		l.Debug().
			Str("status", resp.Status).
			Int("status_code", resp.StatusCode).
			Msg("http request returned a non-200 status code")
		return resp, err
	}

	l.Trace().Msg("finished http request successfully")
	return resp, err
}
