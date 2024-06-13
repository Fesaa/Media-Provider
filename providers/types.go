package providers

import "github.com/Fesaa/Media-Provider/config"

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
	ImageUrl    string `json:"ImageUrl"`
	RefUrl      string `json:"RefUrl"`
}

type SearchRequest struct {
	Provider config.SearchProvider `json:"provider,omitempty"`
	Query    string                `json:"query"`
	Category string                `json:"category,omitempty"`
	SortBy   string                `json:"sort_by,omitempty"`
	Filter   string                `json:"filter,omitempty"`
}

type searchProvider interface {
	Search(request SearchRequest) ([]TorrentInfo, error)
}
