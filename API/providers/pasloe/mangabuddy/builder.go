package mangabuddy

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/rs/zerolog"
)

const (
	domain = "https://mangabuddy.com"
)

type SearchOptions struct {
	Query   string
	Genres  []string
	Status  string
	OrderBy string
}

type Builder struct {
	log        zerolog.Logger
	httpClient *menou.Client
	ps         publication.Client
	repository Repository
	cache      services.CacheService
}

func (b *Builder) Provider() models.Provider {
	return models.MANGA_BUDDY
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(ctx context.Context, t []payload.Info) []payload.Info {
	return t
}

func (b *Builder) Transform(ctx context.Context, request payload.SearchRequest) SearchOptions {
	so := SearchOptions{}

	so.Query = request.Query

	genres, ok := request.Modifiers["genres"]
	if ok {
		so.Genres = genres
	}

	orderBy, ok := request.Modifiers["sort"]
	if ok && len(orderBy) > 0 {
		so.OrderBy = orderBy[0]
	}

	status, ok := request.Modifiers["status"]
	if ok && len(status) > 0 {
		so.Status = status[0]
	}

	return so

}

func (b *Builder) Search(ctx context.Context, s SearchOptions) ([]payload.Info, error) {
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
				Key:      publication.DownloadOneShotKey,
				FormType: payload.SWITCH,
			},
			{
				Key:      publication.TitleOverride,
				Advanced: true,
				FormType: payload.TEXT,
			},
			{
				Key:      publication.AssignEmptyVolumes,
				Advanced: true,
				FormType: payload.SWITCH,
			},
			{
				Key:      publication.ScanlationGroupKey,
				Advanced: true,
				FormType: payload.TEXT,
			},
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ps
}

func NewBuilder(log zerolog.Logger, httpClient *menou.Client, ps publication.Client, repository Repository, cache services.CacheService) *Builder {
	return &Builder{
		log:        log.With().Str("handler", "mangabuddy-provider").Logger(),
		httpClient: httpClient,
		ps:         ps,
		repository: repository,
		cache:      cache,
	}
}
