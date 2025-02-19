package mangadex

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"net/http"
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

type CoverFactory func(volume string) ([]byte, bool)

var defaultCoverFactory CoverFactory = func(volume string) ([]byte, bool) { return nil, false }

func (m *manga) getCoverBytes(fileName string) ([]byte, error) {
	url := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.512.jpg", m.id, fileName)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			m.Log.Warn().Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m *manga) getCoverFactoryLang(coverResp *MangaCoverResponse) CoverFactory {
	if len(coverResp.Data) == 0 {
		return defaultCoverFactory
	}

	firstId := coverResp.Data[0].Id

	covers := make(map[string][]byte)
	var defaultBytes []byte

	processCover := func(cover MangaCoverData) error {
		coverBytes, err := m.getCoverBytes(cover.Attributes.FileName)
		if err != nil {
			m.Log.Err(err).Str("fileName", cover.Attributes.FileName).Msg("Failed to get cover")
			return err
		}
		covers[cover.Attributes.Volume] = coverBytes
		if cover.Id == firstId {
			defaultBytes = coverBytes
		}
		return nil
	}

	coversByLang := utils.GroupBy(coverResp.Data, func(v MangaCoverData) string {
		return v.Attributes.Locale
	})

	if wanted, ok := coversByLang[m.language]; ok {
		for _, cover := range wanted {
			if err := processCover(cover); err != nil {
				return defaultCoverFactory
			}
		}
	}

	for _, cover := range coverResp.Data {
		if cover.Attributes.Locale == m.language {
			continue // Already gone over
		}

		if _, ok := covers[cover.Attributes.Volume]; !ok {
			if err := processCover(cover); err != nil {
				return defaultCoverFactory
			}
		}
	}

	return func(volume string) ([]byte, bool) {
		url, ok := covers[volume]
		if !ok && len(defaultBytes) != 0 {
			return defaultBytes, true
		}
		if !ok {
			return nil, false
		}

		return url, true
	}
}
