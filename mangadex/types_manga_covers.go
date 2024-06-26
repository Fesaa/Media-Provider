package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
)

type MangaCoverResponse MangaDexResponse[[]MangaCoverData]

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

func (m *MangaCoverResponse) GetUrlsPerVolume(mangaId string) map[string]string {
	out := make(map[string]string)
	log.Debug("getting cover urls for manga", "mangaId", mangaId, "amount", len(m.Data))

	coverUrl := func(fileName string) string {
		return fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", mangaId, fileName)
	}

	for _, cover := range m.Data {
		url := coverUrl(cover.Attributes.FileName)
		log.Trace("setting cover url for volume", "mangaId", mangaId, "volume", cover.Attributes.Volume, "url", url)
		out[cover.Attributes.Volume] = url
	}

	return out
}
