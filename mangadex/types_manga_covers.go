package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
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
	log.Trace("getting cover urls for manga", "mangaId", mangaId, "amount", len(m.Data))

	coverUrl := func(fileName string) string {
		return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mangaId, fileName)
	}

	var defaultCover string
	if len(m.Data) > 0 {
		// Set the first cover as the default cover, this way a manga always has a cover
		// Even if it's a bit wrong
		defaultCover = coverUrl(m.Data[0].Attributes.FileName)
	}

	for _, cover := range m.Data {
		url := coverUrl(cover.Attributes.FileName)
		log.Trace("setting cover url for volume", "mangaId", mangaId, "volume", cover.Attributes.Volume, "url", url)
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
