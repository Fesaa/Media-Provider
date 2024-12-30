package payload

import (
	"github.com/Fesaa/Media-Provider/db/models"
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
	Provider    models.Provider `json:"Provider"`
}

type InfoTag struct {
	Name  string `json:"Name"`
	Value any    `json:"Value"`
}

func Of(name string, value any) InfoTag {
	return InfoTag{
		Name:  name,
		Value: value,
	}
}
