package limetorrents

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/rs/zerolog"
	"net/http"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *http.Client
	ys         yoitsu.Yoitsu
}

func (b *Builder) Provider() models.Provider {
	return models.LIME
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(torrents []SearchResult) []payload.Info {
	torrentsInfo := make([]payload.Info, len(torrents))
	for i, t := range torrents {
		torrentsInfo[i] = payload.Info{
			Name:        t.Name,
			Description: "",
			Size:        t.Size,
			Tags: []payload.InfoTag{
				payload.Of("Date", t.Added),
				payload.Of("Seeders", t.Seed),
				payload.Of("Leechers", t.Leach),
			},
			Link:     t.Url,
			InfoHash: t.Hash,
			ImageUrl: "",
			RefUrl:   t.PageUrl,
			Provider: models.LIME,
		}
	}
	return torrentsInfo
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	categories, ok := s.Modifiers["categories"]
	var category string
	if ok && len(categories) > 0 {
		category = categories[0]
	}
	return SearchOptions{
		Category: ConvertCategory(category),
		Query:    s.Query,
		Page:     1,
	}
}

func (b *Builder) Download(request payload.DownloadRequest) error {
	_, err := b.ys.AddDownload(request)
	return err
}

func (b *Builder) Stop(request payload.StopRequest) error {
	return b.ys.RemoveDownload(request)
}

func (b *Builder) DownloadMetadata() payload.DownloadMetadata {
	return payload.DownloadMetadata{}
}

func NewBuilder(log zerolog.Logger, httpClient *http.Client, ys yoitsu.Yoitsu) *Builder {
	return &Builder{log.With().Str("handler", "limetorrents-provider").Logger(), httpClient, ys}
}
