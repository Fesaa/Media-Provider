package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
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

func (m *MangaCoverResponse) GetCoverFactoryLang(lang string, mangaId string) CoverFactory {
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

	coversByLang := utils.GroupBy(m.Data, func(v MangaCoverData) string {
		return v.Attributes.Locale
	})

	if wanted, ok := coversByLang[lang]; ok {
		for _, cover := range wanted {
			url := coverUrl(cover.Attributes.FileName)
			covers[cover.Attributes.Volume] = url
		}
	}

	for _, cover := range m.Data {
		if _, ok := covers[cover.Attributes.Volume]; !ok {
			url := coverUrl(cover.Attributes.FileName)
			covers[cover.Attributes.Volume] = url
		}
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
