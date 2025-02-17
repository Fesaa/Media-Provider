package yts

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"net/http"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *http.Client
	ys         yoitsu.Yoitsu
}

func (b *Builder) Provider() models.Provider {
	return models.YTS
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(data *SearchResult) []payload.Info {
	movies := data.Data.Movies
	torrents := make([]payload.Info, len(movies))
	for i, movie := range movies {
		var torrent *Torrent = nil
		for _, t := range movie.Torrents {
			if t.Quality == "1080p" {
				torrent = &t
				break
			}
		}
		if torrent == nil {
			torrent = &movie.Torrents[0]
		}

		torrents[i] = payload.Info{
			Name:        movie.Title,
			Description: movie.DescriptionFull,
			Size:        torrent.Size,
			Tags: []payload.InfoTag{
				payload.Of("Date", utils.Stringify(movie.Year)),
				payload.Of("Seeders", utils.Stringify(torrent.Seeds)),
				payload.Of("Leechers", utils.Stringify(torrent.Peers)),
			},
			Link:     torrent.Url,
			InfoHash: torrent.Hash,
			ImageUrl: movie.MediumCoverImage,
			RefUrl:   movie.Url,
			Provider: models.YTS,
		}
	}
	return torrents
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	y := SearchOptions{}
	y.Query = s.Query
	sortBys, ok := s.Modifiers["sortBys"]
	if ok && len(sortBys) > 0 {
		y.SortBy = sortBys[0]
	}
	y.Page = 1
	return y
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

func NewBuilder(log zerolog.Logger, httpClient *http.Client, ys yoitsu.Yoitsu) *Builder {
	return &Builder{log.With().Str("handler", "yts-provider").Logger(), httpClient, ys}
}
