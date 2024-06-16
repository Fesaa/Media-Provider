package providers

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
)

type Info struct {
	Name        string          `json:"Name"`
	Description string          `json:"Description"`
	Date        string          `json:"Date"`
	Size        string          `json:"Size"`
	Seeders     string          `json:"Seeders"`
	Leechers    string          `json:"Leechers"`
	Downloads   string          `json:"Downloads"`
	Link        string          `json:"Link"`
	InfoHash    string          `json:"InfoHash"`
	ImageUrl    string          `json:"ImageUrl"`
	RefUrl      string          `json:"RefUrl"`
	Provider    config.Provider `json:"Provider"`
}

type provider interface {
	Search(payload.SearchRequest) ([]Info, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}
