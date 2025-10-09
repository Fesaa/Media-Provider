package dynasty

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *menou.Client
	ps         publication.Client
	repository Repository
}

func (b *Builder) Provider() models.Provider {
	return models.DYNASTY
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(ctx context.Context, mangas []SearchData) []payload.Info {
	if mangas == nil {
		return []payload.Info{}
	}

	info := make([]payload.Info, len(mangas))
	for i, manga := range mangas {
		info[i] = payload.Info{
			Name: manga.Title,
			Tags: utils.Map(manga.Tags, func(t publication.Tag) payload.InfoTag {
				return payload.Of(t.Value, "")
			}),
			InfoHash: manga.Id,
			RefUrl:   manga.RefUrl(),
			Provider: models.DYNASTY,
		}
	}

	return info
}

func (b *Builder) Transform(ctx context.Context, s payload.SearchRequest) SearchOptions {
	return SearchOptions{
		Query: s.Query,
	}
}

func (b *Builder) Search(ctx context.Context, s SearchOptions) ([]SearchData, error) {
	return b.repository.SearchSeries(ctx, s)
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{
		Definitions: []payload.DownloadMetadataDefinition{
			{
				Key:      publication.DownloadOneShotKey,
				FormType: payload.SWITCH,
			},
			{
				Key:      publication.IncludeNotMatchedTagsKey,
				Advanced: true,
				FormType: payload.SWITCH,
			},
			{
				Key:           publication.IncludeCover,
				FormType:      payload.SWITCH,
				DefaultOption: "true",
			},
			{
				Key:      publication.TitleOverride,
				Advanced: true,
				FormType: payload.TEXT,
			},
			{
				Key:      publication.SkipVolumeWithoutChapter,
				Advanced: true,
				FormType: payload.SWITCH,
			},
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ps
}

func NewBuilder(log zerolog.Logger, httpClient *menou.Client, ps publication.Client, repository Repository) *Builder {
	return &Builder{log.With().Str("handler", "dynasty-provider").Logger(),
		httpClient, ps, repository}
}
