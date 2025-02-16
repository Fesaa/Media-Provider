package mangadex

import (
	"fmt"
)

type MangaCoverResponse Response[[]MangaCoverData]

type MangaCoverData struct {
	Id            string               `json:"id"`
	Type          string               `json:"type"`
	Attributes    MangaCoverAttributes `json:"attributes"`
	Relationships []Relationship       `json:"relationships"`
}

type MangaCoverAttributes struct {
	Description string `json:"description"`
	Volume      string `json:"volume"`
	FileName    string `json:"fileName"`
	Locale      string `json:"locale"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Version     int    `json:"version"`
}

type CoverFactory func(volume string) (string, bool)

func (m *MangaCoverResponse) GetCoverFactory(mangaId string) CoverFactory {
	covers := make(map[string]string)

	coverUrl := func(fileName string) string {
		return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.512.jpg", mangaId, fileName)
	}

	var defaultCover string
	if len(m.Data) > 0 {
		// Set the first cover as the default cover, this way a manga always has a cover
		// Even if it's a bit wrong
		defaultCover = coverUrl(m.Data[0].Attributes.FileName)
	}

	for _, cover := range m.Data {
		url := coverUrl(cover.Attributes.FileName)
		covers[cover.Attributes.Volume] = url
	}

	return func(volume string) (string, bool) {
		url, ok := covers[volume]
		if !ok && defaultCover != "" {
			return defaultCover, true
		}
		if !ok {
			return "", false
		}

		return url, true
	}
}
