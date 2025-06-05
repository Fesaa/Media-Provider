package menou

import (
	"github.com/rs/zerolog"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
	log zerolog.Logger
}

func NewWithRetry(log zerolog.Logger) *Client {
	return &Client{
		&http.Client{
			Transport: &loggingTransport{
				Transport: &retryer{
					log: log.With().Str("handler", "httpClient-retryer").Logger(),
				},
				log: log.With().Str("handler", "httpClient").Logger(),
			},
			Timeout: time.Second * 30,
		},
		log.With().Str("handler", "menou").Logger(),
	}
}

func New(log zerolog.Logger) *Client {
	return &Client{
		&http.Client{
			Transport: &loggingTransport{
				log: log.With().Str("handler", "httpClient").Logger(),
			},
			Timeout: time.Second * 30,
		},
		log.With().Str("handler", "menou").Logger(),
	}
}
