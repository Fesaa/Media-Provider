package providers

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/rs/zerolog"
	"sync"
)

type ContentProvider struct {
	lock      *sync.Mutex
	providers map[models.Provider]Provider
	log       zerolog.Logger
}

type ProviderBuilder[T, S any] interface {
	Provider() models.Provider
	Logger() zerolog.Logger
	Normalize(T) []payload.Info
	Transform(payload.SearchRequest) S
	Search(S) (T, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}

type Provider interface {
	Search(payload.SearchRequest) ([]payload.Info, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}
