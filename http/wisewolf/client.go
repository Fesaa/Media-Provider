package wisewolf

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

func NewWithRetry(log zerolog.Logger) *http.Client {
	return &http.Client{
		Transport: &loggingTransport{
			Transport: &retryer{
				log: log.With().Str("handler", "httpClient-retryer").Logger(),
			},
			log: log.With().Str("handler", "httpClient").Logger(),
		},
		Timeout: time.Second * 30,
	}
}

func New(log zerolog.Logger) *http.Client {
	return &http.Client{
		Transport: &loggingTransport{
			log: log.With().Str("handler", "httpClient").Logger(),
		},
		Timeout: time.Second * 30,
	}
}
