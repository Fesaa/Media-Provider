package mangadex

import "github.com/Fesaa/Media-Provider/config"

type MangadexClient interface {
	Download(id string, baseDir string) error
	RemoveDownload(id string, deleteFiles bool) error
	GetBaseDir() string
	GetCurrentManga() Manga
}

type Manga interface {
	Title() string
	Id() string
	GetBaseDir() string
	Cancel()
	WaitForInfoAndDownload()
	GetInfo() config.Info
}

type SearchOptions struct {
	Query                  string
	IncludedTags           []string
	ExcludedTags           []string
	Status                 []string
	ContentRating          []string
	PublicationDemographic []string
	SkipNotFoundTags       bool
}

type MangaDexResponse[T any] struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     T      `json:"data"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Total    int    `json:"total"`
}

type Relationship struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Related string `json:"related"`
}
