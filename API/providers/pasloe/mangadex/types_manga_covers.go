package mangadex

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
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
	Data MangaCoverData
}

type CoverFactory func(volume string) (*Cover, bool)

var defaultCoverFactory CoverFactory = func(volume string) (*Cover, bool) { return nil, false }

func (m *manga) CustomizePreDownloadHook(ctx context.Context) {
	if !m.Req.GetBool(core.IncludeCover, true) {
		m.coverFactory = defaultCoverFactory
		return
	}

	covers, err := m.repository.GetCoverImages(ctx, m.id)
	if err != nil || covers == nil {
		m.Log.Warn().Err(err).Msg("error while loading manga coverFactory, ignoring")
		m.coverFactory = defaultCoverFactory
		return
	}

	m.coverFactory = m.getCoverFactoryLang(ctx, covers)
}

func (m *manga) getCoverBytes(ctx context.Context, fileName string) ([]byte, error) {
	url := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.512.jpg", m.id, fileName)

	if b, err := m.cache.GetWithContext(ctx, url); err == nil && b != nil {
		return b, nil
	}

	resp, err := m.httpClient.GetWithContext(ctx, url)
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

	if err = m.cache.SetWithContext(ctx, url, data, m.cache.DefaultExpiration()); err != nil {
		m.Log.Warn().Err(err).Msg("Failed to cache response")
	}

	return data, nil
}

func (m *manga) getCoverFactoryLang(ctx context.Context, coverResp *MangaCoverResponse) CoverFactory {
	if len(coverResp.Data) == 0 {
		return defaultCoverFactory
	}

	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeCovers)
	defer span.End()

	covers, firstCover, lastCover := m.processCovers(ctx, coverResp)
	return m.constructCoverFactory(covers, firstCover, lastCover)
}

func (m *manga) processCovers(ctx context.Context, coverResp *MangaCoverResponse) (map[string]*Cover, *Cover, *Cover) {
	covers := make(map[string]*Cover)
	var firstCover, lastCover, firstCoverLang, lastCoverLang *Cover

	for _, cover := range coverResp.Data {
		// Don't download non-matching locale's again if a cover is already present
		if _, ok := covers[cover.Attributes.Volume]; ok && cover.Attributes.Locale != m.language {
			continue
		}

		coverBytes, err := m.getCoverBytes(ctx, cover.Attributes.FileName)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, nil, nil
			}
			m.Log.Err(err).Str("fileName", cover.Attributes.FileName).Msg("Failed to get cover")
			continue
		}

		// Cover is too small for Kavita. Looks weird
		if !m.imageService.IsCover(coverBytes) {
			m.Log.Debug().Str("id", cover.Id).Str("volume", cover.Attributes.Volume).
				Str("desc", cover.Attributes.Description).Msg("cover failed the ImageService.IsCover check. not using")
			continue
		}

		_cover := &Cover{
			Data: cover,
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
	return covers, firstCover, lastCover
}

func (m *manga) constructCoverFactory(covers map[string]*Cover, firstCover, lastCover *Cover) CoverFactory {
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
		return firstCover
	}()

	return func(volume string) (*Cover, bool) {
		cover, ok := covers[volume]
		if !ok && fallbackCover != nil {
			return fallbackCover, true
		}
		if !ok {
			return nil, false
		}

		return cover, true
	}
}
