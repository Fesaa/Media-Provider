package providers

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/subsplease"
	"github.com/Fesaa/Media-Provider/yts"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/irevenko/go-nyaa/types"
	"log/slog"
)

func subsPleaseNormalizer(torrents subsplease.SearchResult) []TorrentInfo {
	if torrents == nil {
		return []TorrentInfo{}
	}

	torrentsInfo := make([]TorrentInfo, 0)
	for name, data := range torrents {
		if len(data.Downloads) == 0 {
			continue
		}
		download := data.Downloads[len(data.Downloads)-1]
		m, err := metainfo.ParseMagnetUri(download.Magnet)
		if err != nil {
			slog.Debug("Couldn't parse magnet uri", "error", err, "info", fmt.Sprintf("%+v", data))
		}
		torrentsInfo = append(torrentsInfo, TorrentInfo{
			Name:        name,
			Description: "",
			Date:        data.ReleaseDate,
			Size:        "",
			Seeders:     "",
			Leechers:    "",
			Downloads:   "",
			Link:        "",
			InfoHash:    m.InfoHash.HexString(),
			ImageUrl:    data.ImageURL,
			RefUrl:      data.ReferenceURL(),
		})
	}
	return torrentsInfo
}

func limeNormalizer(torrents []limetorrents.SearchResult) []TorrentInfo {
	torrentsInfo := make([]TorrentInfo, len(torrents))
	for i, t := range torrents {
		torrentsInfo[i] = TorrentInfo{
			Name:        t.Name,
			Description: "",
			Date:        t.Added,
			Size:        t.Size,
			Seeders:     t.Seed,
			Leechers:    t.Leach,
			Downloads:   "N/A",
			Link:        t.Url,
			InfoHash:    t.Hash,
			ImageUrl:    "",
			RefUrl:      t.PageUrl,
		}
	}
	return torrentsInfo
}

func ytsNormalizer(data *yts.SearchResult) []TorrentInfo {
	movies := data.Data.Movies
	torrents := make([]TorrentInfo, len(movies))
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

		torrents[i] = TorrentInfo{
			Name:        movie.Title,
			Description: movie.DescriptionFull,
			Date:        stringify(movie.Year),
			Size:        torrent.Size,
			Seeders:     stringify(torrent.Seeds),
			Leechers:    stringify(torrent.Peers),
			Downloads:   "",
			Link:        torrent.Url,
			InfoHash:    torrent.Hash,
			ImageUrl:    movie.MediumCoverImage,
			RefUrl:      movie.Url,
		}
	}
	return torrents
}

func nyaaNormalizer(torrents []types.Torrent) []TorrentInfo {
	torrentsInfo := make([]TorrentInfo, len(torrents))
	for i, t := range torrents {
		torrentsInfo[i] = TorrentInfo{
			Name:        t.Name,
			Description: t.Description,
			Date:        t.Date,
			Size:        t.Size,
			Seeders:     t.Seeders,
			Leechers:    t.Leechers,
			Downloads:   t.Downloads,
			Link:        t.Link,
			InfoHash:    t.InfoHash,
			ImageUrl:    "",
			RefUrl:      t.GUID,
		}
	}
	return torrentsInfo
}

func stringify(i int) string {
	return fmt.Sprintf("%d", i)
}
