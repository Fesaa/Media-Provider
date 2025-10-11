package webtoon

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type Builder struct {
	log        zerolog.Logger
	repository Repository
	ps         publication.Client
}

func (b *Builder) Provider() models.Provider {
	return models.WEBTOON
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(ctx context.Context, webtoons []SearchData) []payload.Info {
	return utils.MaybeMap(webtoons, func(w SearchData) (payload.Info, bool) {
		return payload.Info{
			Name: w.Name,
			Tags: []payload.InfoTag{
				payload.Of("Genre", w.Genre),
				payload.Of("Readers", w.ReadCount),
			},
			InfoHash: w.Id,
			ImageUrl: w.ProxiedImage(),
			RefUrl:   w.Url(),
			Provider: models.WEBTOON,
		}, true
	})
}

func (b *Builder) Transform(ctx context.Context, s payload.SearchRequest) SearchOptions {
	return SearchOptions{
		Query: s.Query,
	}
}

func (b *Builder) Search(ctx context.Context, s SearchOptions) ([]SearchData, error) {
	return b.repository.Search(ctx, s)
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{
		Definitions: []payload.DownloadMetadataDefinition{
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
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ps
}

func NewBuilder(log zerolog.Logger, ps publication.Client, repository Repository) *Builder {
	return &Builder{
		log:        log.With().Str("handler", "webtoon-provider").Logger(),
		repository: repository,
		ps:         ps,
	}
}
