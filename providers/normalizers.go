package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/webtoon"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/irevenko/go-nyaa/types"
	"strconv"
)

func mangadexNormalizer(mangas *mangadex.MangaSearchResponse) []Info {
	if mangas == nil {
		return []Info{}
	}

	info := make([]Info, 0)
	for _, data := range mangas.Data {
		enTitle := data.Attributes.EnTitle()
		if enTitle == "" {
			continue
		}

		info = append(info, Info{
			Name:        enTitle,
			Description: data.Attributes.EnDescription(),
			Size:        config.OrDefault(data.Attributes.LastVolume, "unknown") + " Volumes",
			Tags: []InfoTag{
				of("Date", strconv.Itoa(data.Attributes.Year)),
			},
			InfoHash: data.Id,
			RefUrl:   data.RefURL(),
			Provider: config.MANGADEX,
			ImageUrl: data.CoverURL(),
		})
	}

	return info
}

func subsPleaseNormalizer(torrents subsplease.SearchResult) []Info {
	if torrents == nil {
		return []Info{}
	}

	torrentsInfo := make([]Info, 0)
	for name, data := range torrents {
		if len(data.Downloads) == 0 {
			continue
		}
		download := data.Downloads[len(data.Downloads)-1]
		m, err := metainfo.ParseMagnetUri(download.Magnet)
		if err != nil {
			log.Warn("couldn't parse magnet uri", "error", err, "magnet", download.Magnet)
			continue
		}
		torrentsInfo = append(torrentsInfo, Info{
			Name: name,
			Size: "Unknown",
			Tags: []InfoTag{
				of("Date", data.ReleaseDate),
			},
			InfoHash: m.InfoHash.HexString(),
			ImageUrl: data.ImageUrl(),
			RefUrl:   data.ReferenceURL(),
			Provider: config.SUBSPLEASE,
		})
	}
	return torrentsInfo
}

func limeNormalizer(torrents []limetorrents.SearchResult) []Info {
	torrentsInfo := make([]Info, len(torrents))
	for i, t := range torrents {
		torrentsInfo[i] = Info{
			Name:        t.Name,
			Description: "",
			Size:        t.Size,
			Tags: []InfoTag{
				of("Date", t.Added),
				of("Seeders", t.Seed),
				of("Leechers", t.Leach),
			},
			Link:     t.Url,
			InfoHash: t.Hash,
			ImageUrl: "",
			RefUrl:   t.PageUrl,
			Provider: config.LIME,
		}
	}
	return torrentsInfo
}

func ytsNormalizer(data *yts.SearchResult) []Info {
	movies := data.Data.Movies
	torrents := make([]Info, len(movies))
	for i, movie := range movies {
		var torrent *yts.Torrent = nil
		for _, t := range movie.Torrents {
			if t.Quality == "1080p" {
				torrent = &t
				break
			}
		}
		if torrent == nil {
			torrent = &movie.Torrents[0]
		}

		torrents[i] = Info{
			Name:        movie.Title,
			Description: movie.DescriptionFull,
			Size:        torrent.Size,
			Tags: []InfoTag{
				of("Date", stringify(movie.Year)),
				of("Seeders", stringify(torrent.Seeds)),
				of("Leechers", stringify(torrent.Peers)),
			},
			Link:     torrent.Url,
			InfoHash: torrent.Hash,
			ImageUrl: movie.MediumCoverImage,
			RefUrl:   movie.Url,
			Provider: config.YTS,
		}
	}
	return torrents
}

func nyaaNormalizer(provider config.Provider) responseNormalizerFunc[[]types.Torrent] {
	return func(torrents []types.Torrent) []Info {
		torrentsInfo := make([]Info, len(torrents))
		for i, t := range torrents {
			torrentsInfo[i] = Info{
				Name:        t.Name,
				Description: "", // The description passed here, is some raw html nonsense. Don't use it
				Size:        t.Size,
				Tags: []InfoTag{
					of("Date", t.Date),
					of("Seeders", t.Seeders),
					of("Leechers", t.Leechers),
					of("Downloads", t.Downloads),
				},
				Link:     t.Link,
				InfoHash: t.InfoHash,
				ImageUrl: "",
				RefUrl:   t.GUID,
				Provider: provider,
			}
		}
		return torrentsInfo
	}
}

func webtoonNormalizer(webtoons []webtoon.SearchData) []Info {
	return utils.MaybeMap(webtoons, func(w webtoon.SearchData) (Info, bool) {
		return Info{
			Name: w.Name,
			Tags: []InfoTag{
				of("Genre", w.Genre),
				of("Readers", w.ReadCount),
			},
			InfoHash: stringify(w.Id),
			ImageUrl: w.ProxiedImage(),
			RefUrl:   w.Url(),
			Provider: config.WEBTOON,
		}, true
	})
}

func stringify(i int) string {
	return fmt.Sprintf("%d", i)
}
