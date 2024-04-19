package routes

import (
	"fmt"

	"github.com/Fesaa/Media-Provider/limetorrents"
	"github.com/Fesaa/Media-Provider/yts"
)

type TorrentInfo struct {
	Name        string `json:"Name"`
	Description string `json:"Description"`
	Date        string `json:"Date"`
	Size        string `json:"Size"`
	Seeders     string `json:"Seeders"`
	Leechers    string `json:"Leechers"`
	Downloads   string `json:"Downloads"`
	Link        string `json:"Link"`
	InfoHash    string `json:"InfoHash"`
}

func fromLime(torrents []limetorrents.SearchResult) []TorrentInfo {
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
		}
	}
	return torrentsInfo
}

func fromYTS(movies []yts.YTSMovie) []TorrentInfo {
	torrents := make([]TorrentInfo, len(movies))
	for i, movie := range movies {
		var torrent *yts.YTSTorrent = nil
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
		}
	}
	return torrents
}

func stringify(i int) string {
	return fmt.Sprintf("%d", i)
}
