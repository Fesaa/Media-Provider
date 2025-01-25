package webtoon

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
)

type Builder struct {
	log        zerolog.Logger
	repository Repository
	ps         api.Client
}

func (b *Builder) Provider() models.Provider {
	return models.WEBTOON
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(webtoons []SearchData) []payload.Info {
	return utils.MaybeMap(webtoons, func(w SearchData) (payload.Info, bool) {
		return payload.Info{
			Name: w.Name,
			Tags: []payload.InfoTag{
				payload.Of("Genre", w.Genre),
				payload.Of("Readers", w.ReadCount),
			},
			InfoHash: utils.Stringify(w.Id),
			ImageUrl: w.ProxiedImage(),
			RefUrl:   w.Url(),
			Provider: models.WEBTOON,
		}, true
	})
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	return SearchOptions{
		Query: s.Query,
	}
}

func (b *Builder) Search(s SearchOptions) ([]SearchData, error) {
	return b.repository.Search(s)
}

func (b *Builder) Download(request payload.DownloadRequest) error {
	return b.ps.Download(request)
}

func (b *Builder) Stop(request payload.StopRequest) error {
	return b.ps.RemoveDownload(request)
}

func NewBuilder(log zerolog.Logger, ps api.Client, repository Repository) *Builder {
	return &Builder{
		log:        log.With().Str("handler", "webtoon-provider").Logger(),
		repository: repository,
		ps:         ps,
	}
}
