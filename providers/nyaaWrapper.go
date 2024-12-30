package providers

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/irevenko/go-nyaa/nyaa"
	"github.com/irevenko/go-nyaa/types"
	"github.com/rs/zerolog"
	"net/http"
	"net/url"
)

type NyaaBuilder struct {
	log        zerolog.Logger
	httpClient *http.Client
	ys         yoitsu.Yoitsu
}

func (b *NyaaBuilder) Provider() models.Provider {
	return models.NYAA
}

func (b *NyaaBuilder) Logger() zerolog.Logger {
	return b.log
}

func (b *NyaaBuilder) Normalize(torrents []types.Torrent) []payload.Info {
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

func (b *NyaaBuilder) Transform(s payload.SearchRequest) nyaa.SearchOptions {
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

func (b *NyaaBuilder) Search(opts nyaa.SearchOptions) ([]types.Torrent, error) {
	search, err := nyaa.Search(opts)
	if err != nil {
		return nil, err
	}
	return search, nil
}

func (b *NyaaBuilder) Download(request payload.DownloadRequest) error {
	_, err := b.ys.AddDownload(request)
	return err
}

func (b *NyaaBuilder) Stop(request payload.StopRequest) error {
	return b.ys.RemoveDownload(request)
}

func NewNyaaBuilder(log zerolog.Logger, httpClient *http.Client, ys yoitsu.Yoitsu) *NyaaBuilder {
	return &NyaaBuilder{log.With().Str("handler", "nyaa-provider").Logger(), httpClient, ys}
}
