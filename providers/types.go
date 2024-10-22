package providers

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/payload"
)

type Info struct {
	Name        string          `json:"Name"`
	Description string          `json:"Description"`
	Size        string          `json:"Size"`
	Tags        []InfoTag       `json:"Tags"`
	Link        string          `json:"Link"`
	InfoHash    string          `json:"InfoHash"`
	ImageUrl    string          `json:"ImageUrl"`
	RefUrl      string          `json:"RefUrl"`
	Provider    models.Provider `json:"Providers"`
}

type InfoTag struct {
	Name  string `json:"Name"`
	Value any    `json:"Value"`
}

func of(name string, value any) InfoTag {
	return InfoTag{
		Name:  name,
		Value: value,
	}
}

type provider interface {
	Search(payload.SearchRequest) ([]Info, error)
	Download(payload.DownloadRequest) error
	Stop(payload.StopRequest) error
}
