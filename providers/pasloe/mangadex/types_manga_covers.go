package mangadex

import (
	"errors"
	"fmt"
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

type Cover = []byte

type CoverFactory func(volume string) (Cover, bool)

var defaultCoverFactory CoverFactory = func(volume string) (Cover, bool) { return nil, false }

func (m *manga) getCoverBytes(fileName string) (Cover, error) {
	url := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.512.jpg", m.id, fileName)
	resp, err := m.httpClient.Get(url)
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

	covers := make(map[string]Cover)
	var defaultCover, defaultCoverLang Cover

	for _, cover := range coverResp.Data {
		// Don't download non-matching locale's again if a cover is already present
		if _, ok := covers[cover.Attributes.Volume]; ok && cover.Attributes.Locale != m.language {
			continue
		}

		coverBytes, err := m.getCoverBytes(cover.Attributes.FileName)
		if err != nil {
			m.Log.Err(err).Str("fileName", cover.Attributes.FileName).Msg("Failed to get cover")
			continue
		}

		// Cover is too small for Kavita. Looks weird
		if !m.imageService.IsCover(coverBytes) {
			m.Log.Trace().Str("id", cover.Id).Str("volume", cover.Attributes.Volume).
				Str("desc", cover.Attributes.Description).Msg("cover failed the ImageService.IsCover check. not using")
			continue
		}

		if cover.Attributes.Locale == m.language {
			covers[cover.Attributes.Volume] = coverBytes
			if len(defaultCoverLang) == 0 {
				defaultCoverLang = coverBytes
			}
		} else if _, ok := covers[cover.Attributes.Volume]; !ok {
			covers[cover.Attributes.Volume] = coverBytes
		}

		if len(defaultCover) == 0 {
			defaultCover = coverBytes
		}
	}

	if len(defaultCoverLang) > 0 {
		defaultCover = defaultCoverLang
	}

	return func(volume string) (Cover, bool) {
		url, ok := covers[volume]
		if !ok && len(defaultCover) != 0 {
			return defaultCover, true
		}
		if !ok {
			return nil, false
		}

		return url, true
	}
}
