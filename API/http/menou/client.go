package menou

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var DefaultClient = &Client{
	Client: http.DefaultClient,
	log:    zerolog.Nop(),
}

type Client struct {
	*http.Client
	log zerolog.Logger
}

func NewWithRetry(log zerolog.Logger) *Client {
	baseTransport := &retryer{
		log: log.With().Str("handler", "httpClient-retryer").Logger(),
	}

	logging := &loggingTransport{
		Transport: baseTransport,
		log:       log.With().Str("handler", "httpClient").Logger(),
	}

	traced := otelhttp.NewTransport(logging)

	return &Client{
		&http.Client{
			Transport: traced,
			Timeout:   time.Second * 30,
		},
		log.With().Str("handler", "menou").Logger(),
	}
}

func New(log zerolog.Logger) *Client {
	logging := &loggingTransport{
		log: log.With().Str("handler", "httpClient").Logger(),
	}

	traced := otelhttp.NewTransport(logging)

	return &Client{
		&http.Client{
			Transport: traced,
			Timeout:   time.Second * 30,
		},
		log.With().Str("handler", "menou").Logger(),
	}
}
