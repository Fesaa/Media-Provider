package wisewolf

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

func New(log zerolog.Logger) *http.Client {
	return &http.Client{
		Transport: &loggingTransport{
			log: log.With().Str("handler", "httpClient").Logger(),
		},
		Timeout: time.Second * 30,
	}
}
