package mangadex

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
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

type Cover struct {
	Bytes []byte
	Data  MangaCoverData
}

type CoverFactory func(volume string) (*Cover, bool)

var defaultCoverFactory CoverFactory = func(volume string) (*Cover, bool) { return nil, false }

func (m *manga) getCoverBytes(fileName string) ([]byte, error) {
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

	covers := make(map[string]*Cover)
	var firstCover, lastCover, firstCoverLang, lastCoverLang *Cover

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

		_cover := &Cover{
			Bytes: coverBytes,
			Data:  cover,
		}

		if cover.Attributes.Locale == m.language {
			covers[cover.Attributes.Volume] = _cover
			if firstCoverLang == nil {
				firstCoverLang = _cover
			}
			lastCoverLang = _cover
		} else if _, ok := covers[cover.Attributes.Volume]; !ok {
			covers[cover.Attributes.Volume] = _cover
		}

		if firstCover == nil {
			firstCover = _cover
		}
		lastCover = _cover
	}

	if firstCoverLang != nil {
		firstCover = firstCoverLang
	}
	if lastCoverLang != nil {
		lastCover = lastCoverLang
	}

	fallbackCover := func() *Cover {
		if m.Preference == nil {
			return nil
		}
		switch m.Preference.CoverFallbackMethod {
		case models.CoverFallbackFirst:
			return firstCover
		case models.CoverFallbackLast:
			return lastCover
		case models.CoverFallbackNone:
			return nil
		}
		return nil
	}()

	return func(volume string) (*Cover, bool) {
		url, ok := covers[volume]
		if !ok && fallbackCover != nil && len(fallbackCover.Bytes) != 0 {
			return fallbackCover, true
		}
		if !ok {
			return nil, false
		}

		return url, true
	}
}
