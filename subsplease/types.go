package subsplease

import "fmt"

type SearchResult map[string]TorrentData

type TorrentData struct {
	Time        string     `json:"time"`
	ReleaseDate string     `json:"release_date"`
	Show        string     `json:"show"`
	Episode     string     `json:"episode"`
	Downloads   []Download `json:"downloads"`
	XDCC        string     `json:"xdcc"`
	ImageURL    string     `json:"image_url"`
	Page        string     `json:"page"`
}

func (t TorrentData) ReferenceURL() string {
	return fmt.Sprintf("https://subsplease.org/shows/%s/", t.Page)
}

type Download struct {
	Res    string `json:"res"`
	Magnet string `json:"magnet"`
}
