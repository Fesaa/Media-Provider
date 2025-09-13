package nyaa

import (
	"context"
	"net/url"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/irevenko/go-nyaa/nyaa"
	"github.com/irevenko/go-nyaa/types"
	"github.com/rs/zerolog"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *menou.Client
	ys         yoitsu.Client
}

func (b *Builder) Provider() models.Provider {
	return models.NYAA
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(ctx context.Context, torrents []types.Torrent) []payload.Info {
	torrentsInfo := make([]payload.Info, len(torrents))
	for i, t := range torrents {
		torrentsInfo[i] = payload.Info{
			Name:        t.Name,
			Description: "", // The description passed here, is some raw html nonsense. Don't use it
			Size:        t.Size,
			Tags: []payload.InfoTag{
				payload.Of("Date", t.Date),
				payload.Of("Seeders", t.Seeders),
				payload.Of("Leechers", t.Leechers),
				payload.Of("Downloads", t.Downloads),
			},
			Link:     t.Link,
			InfoHash: t.InfoHash,
			ImageUrl: "",
			RefUrl:   t.GUID,
			Provider: models.NYAA,
		}
	}
	return torrentsInfo
}

func (b *Builder) Transform(ctx context.Context, s payload.SearchRequest) nyaa.SearchOptions {
	so := nyaa.SearchOptions{}
	so.Query = url.QueryEscape(s.Query)
	so.Provider = "nyaa"
	categories, ok := s.Modifiers["categories"]
	if ok && len(categories) > 0 {
		so.Category = categories[0]
	}

	sortBys, ok := s.Modifiers["sortBys"]
	if ok && len(sortBys) > 0 {
		so.SortBy = sortBys[0]
	}

	filters, ok := s.Modifiers["filters"]
	if ok && len(filters) > 0 {
		so.Filter = filters[0]
	}

	return so
}

func (b *Builder) Search(ctx context.Context, opts nyaa.SearchOptions) ([]types.Torrent, error) {
	search, err := nyaa.Search(opts)
	if err != nil {
		return nil, err
	}
	return search, nil
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{
		Definitions: []payload.DownloadMetadataDefinition{
			{
				Key:           "no-sub-dir",
				FormType:      payload.SWITCH,
				DefaultOption: "",
			},
		},
	}
}

func (b *Builder) Client() services.Client {
	return b.ys
}

func NewBuilder(log zerolog.Logger, httpClient *menou.Client, ys yoitsu.Client) *Builder {
	return &Builder{log.With().Str("handler", "nyaa-provider").Logger(), httpClient, ys}
}
