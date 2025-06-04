package bato

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
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
	return models.BATO
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(t []SearchResult) []payload.Info {
	return utils.Map(t, func(t SearchResult) payload.Info {
		return payload.Info{
			Name:     t.Title,
			Tags:     nil,
			ImageUrl: t.ImageUrl,
			InfoHash: t.Id,
			RefUrl:   fmt.Sprintf("%s/title/%s", Domain, t.Id),
			Provider: models.BATO,
		}
	})
}

func (b *Builder) Transform(request payload.SearchRequest) SearchOptions {
	so := SearchOptions{}

	so.Query = request.Query

	genres, ok := request.Modifiers[GenresTag]
	if ok {
		so.Genres = genres
	}

	origLang, ok := request.Modifiers[OriginalLangTag]
	if ok {
		so.OriginalLang = origLang
	}

	transLang, ok := request.Modifiers[TranslatedLangTag]
	if ok {
		so.TranslatedLang = transLang
	}

	workStatus, ok := request.Modifiers[StatusTag]
	if ok {
		so.OriginalWorkStatus = utils.MaybeMap(workStatus, toPublication)
	}

	uploadStatus, ok := request.Modifiers[UploadTag]
	if ok {
		so.BatoUploadStatus = utils.MaybeMap(uploadStatus, toPublication)
	}

	return so

}

func (b *Builder) Search(s SearchOptions) ([]SearchResult, error) {
	return b.repository.Search(context.Background(), s)
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{
		Definitions: []payload.DownloadMetadataDefinition{
			{
				Key:           api.IncludeCover,
				FormType:      payload.SWITCH,
				DefaultOption: "true",
			},
			{
				Key:      api.DownloadOneShotKey,
				FormType: payload.SWITCH,
			},
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ps
}

func NewBuilder(log zerolog.Logger, httpClient *http.Client, ps api.Client, repository Repository) *Builder {
	return &Builder{
		log:        log.With().Str("handler", "dynasty-provider").Logger(),
		httpClient: httpClient,
		ps:         ps,
		repository: repository,
	}
}
