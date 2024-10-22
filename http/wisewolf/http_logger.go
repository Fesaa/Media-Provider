package wisewolf

import (
	"github.com/Fesaa/Media-Provider/log"
	"log/slog"
	"net/http"
	"time"
)

type loggingTransport struct {
	Transport http.RoundTripper
}

func (lt *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if lt.Transport == nil {
		lt.Transport = http.DefaultTransport
	}

	startTime := time.Now()
	resp, err := lt.Transport.RoundTrip(req)
	duration := time.Since(startTime)

	l := log.With(
		slog.String("url", req.URL.String()),
		slog.String("method", req.Method),
		slog.Duration("duration", duration),
	)

	if err != nil {
		l.Trace("http request returned a non-nil error", "err", err)
		return resp, err
	}

	if resp.StatusCode >= 400 {
		l.Debug("http request returned an error status code",
			slog.String("status", resp.Status),
			slog.Int("status_code", resp.StatusCode))
		return resp, err
	}

	l.Trace("finished http request successfully")
	return resp, err
}
