package bato

import (
	"context"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *menou.Client
	ps         core.Client
	repository Repository
	cache      services.CacheService
}

func (b *Builder) Provider() models.Provider {
	return models.BATO
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(t []SearchResult) []payload.Info {
	return utils.Map(t, func(t SearchResult) payload.Info {
		if err := b.cache.Set(t.Id, []byte(t.ImageUrl), time.Hour*24); err != nil {
			b.log.Warn().Err(err).Str("id", t.Id).Msg("failed to cache image")
		}
		return payload.Info{
			Name:     t.Title,
			Tags:     []payload.InfoTag{},
			ImageUrl: fmt.Sprintf("proxy/bato/covers/%s", t.Id),
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
				Key:           core.IncludeCover,
				FormType:      payload.SWITCH,
				DefaultOption: "true",
			},
			{
				Key:      core.DownloadOneShotKey,
				FormType: payload.SWITCH,
			},
			{
				Key:      core.TitleOverride,
				Advanced: true,
				FormType: payload.TEXT,
			},
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ps
}

func NewBuilder(log zerolog.Logger, httpClient *menou.Client, ps core.Client, repository Repository, cache services.CacheService) *Builder {
	return &Builder{
		log:        log.With().Str("handler", "dynasty-provider").Logger(),
		httpClient: httpClient,
		ps:         ps,
		repository: repository,
		cache:      cache,
	}
}
