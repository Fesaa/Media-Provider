package subsplease

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/rs/zerolog"
)

type Builder struct {
	log        zerolog.Logger
	httpClient *menou.Client
	ys         yoitsu.Yoitsu
}

func (b *Builder) Provider() models.Provider {
	return models.SUBSPLEASE
}

func (b *Builder) Logger() zerolog.Logger {
	return b.log
}

func (b *Builder) Normalize(torrents SearchResult) []payload.Info {
	if torrents == nil {
		return []payload.Info{}
	}

	torrentsInfo := make([]payload.Info, 0)
	for name, data := range torrents {
		if len(data.Downloads) == 0 {
			continue
		}
		download := data.Downloads[len(data.Downloads)-1]
		m, err := metainfo.ParseMagnetUri(download.Magnet)
		if err != nil {
			continue
		}
		torrentsInfo = append(torrentsInfo, payload.Info{
			Name: name,
			Size: "Unknown",
			Tags: []payload.InfoTag{
				payload.Of("Date", data.ReleaseDate),
			},
			InfoHash: m.InfoHash.HexString(),
			ImageUrl: data.ImageUrl(),
			RefUrl:   data.ReferenceURL(),
			Provider: models.SUBSPLEASE,
		})
	}
	return torrentsInfo
}

func (b *Builder) Transform(s payload.SearchRequest) SearchOptions {
	return SearchOptions{
		Query: s.Query,
	}
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

func NewBuilder(log zerolog.Logger, httpClient *menou.Client, ys yoitsu.Yoitsu) *Builder {
	return &Builder{log.With().Str("handler", "subsplease-provider").Logger(), httpClient, ys}
}
