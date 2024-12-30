package pasloe

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	utils2 "github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"net/http"
)

func New(cfg *config.Config, container *dig.Container, httpClient *http.Client, log zerolog.Logger) api.Client {
	utils2.Must(container.Invoke(mangadex.Init))
	return newClient(cfg, httpClient, container, log)
}
