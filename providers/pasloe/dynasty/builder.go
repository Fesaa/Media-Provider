package dynasty

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"net/http"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *http.Client
	ps         api.Client
	repository Repository
}

func (b *Builder) Provider() models.Provider {
	return models.DYNASTY
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(mangas []SearchData) []payload.Info {
	if mangas == nil {
		return []payload.Info{}
	}

	info := make([]payload.Info, len(mangas))
	for i, manga := range mangas {
		info[i] = payload.Info{
			Name: manga.Title,
			Tags: utils.Map(manga.Tags, func(t Tag) payload.InfoTag {
				return payload.Of(t.DisplayName, "")
			}),
			InfoHash: manga.Id,
			RefUrl:   manga.RefUrl(),
			Provider: models.DYNASTY,
		}
	}

	return info
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	return SearchOptions{
		Query: s.Query,
	}
}

func (b *Builder) Search(s SearchOptions) ([]SearchData, error) {
	return b.repository.SearchSeries(s)
}

func (b *Builder) Download(request payload.DownloadRequest) error {
	return b.ps.Download(request)
}

func (b *Builder) Stop(request payload.StopRequest) error {
	return b.ps.RemoveDownload(request)
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{
		Definitions: []payload.DownloadMetadataDefinition{
			{
				Title:         "Download OneShots",
				Key:           DownloadOneShotKey,
				FormType:      payload.SWITCH,
				DefaultOption: "false",
			},
		},
	}
}

func NewBuilder(log zerolog.Logger, httpClient *http.Client, ps api.Client, repository Repository) *Builder {
	return &Builder{log.With().Str("handler", "dynasty-provider").Logger(),
		httpClient, ps, repository}
}
