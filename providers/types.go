package providers

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

type SearchProvider string

const (
	NYAA       SearchProvider = "nyaa"
	YTS        SearchProvider = "yts"
	LIME       SearchProvider = "limetorrents"
	SUBSPLEASE SearchProvider = "subsplease"
)

type SearchRequest struct {
	Provider SearchProvider `json:"provider,omitempty"`
	Query    string         `json:"query"`
	Category string         `json:"category,omitempty"`
	SortBy   string         `json:"sort_by,omitempty"`
	Filter   string         `json:"filter,omitempty"`
}

type searchProvider interface {
	Search(request SearchRequest) ([]TorrentInfo, error)
}
