package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/rs/zerolog"
	"net/http"
	"strconv"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *http.Client
	ps         api.Client
	repository *Repository
}

func (b *Builder) Provider() models.Provider {
	return models.MANGADEX
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(mangas *MangaSearchResponse) []payload.Info {
	if mangas == nil {
		return []payload.Info{}
	}

	info := make([]payload.Info, 0)
	for _, data := range mangas.Data {
		enTitle := data.Attributes.EnTitle()
		if enTitle == "" {
			continue
		}

		info = append(info, payload.Info{
			Name:        enTitle,
			Description: data.Attributes.EnDescription(),
			Size: func() string {
				volumes := config.OrDefault(data.Attributes.LastVolume, "unknown")
				chapters := config.OrDefault(data.Attributes.LastChapter, "unknown")
				return fmt.Sprintf("%s Vol. %s Ch.", volumes, chapters)
			}(),
			Tags: []payload.InfoTag{
				payload.Of("Date", strconv.Itoa(data.Attributes.Year)),
			},
			InfoHash: data.Id,
			RefUrl:   data.RefURL(),
			Provider: models.MANGADEX,
			ImageUrl: data.CoverURL(),
		})
	}

	return info
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	return SearchOptions{
		Query: s.Query,
	}
}

func (b *Builder) Search(s SearchOptions) (*MangaSearchResponse, error) {
	return b.repository.SearchManga(s)
}

func (b *Builder) Download(request payload.DownloadRequest) error {
	return b.ps.Download(request)
}

func (b *Builder) Stop(request payload.StopRequest) error {
	return b.ps.RemoveDownload(request)
}

func NewBuilder(log zerolog.Logger, httpClient *http.Client, ps api.Client, repository *Repository) *Builder {
	return &Builder{log.With().Str("handler", "mangadex-provider").Logger(),
		httpClient, ps, repository}
}
