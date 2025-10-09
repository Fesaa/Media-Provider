package mangadex

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/tracing"
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

// CoverFactory returns the url for the given volume
type CoverFactory func(volume string) (string, bool)

var defaultCoverFactory CoverFactory = func(volume string) (string, bool) { return "", false }

func (r *repository) getCoverBytes(ctx context.Context, url string) ([]byte, error) {
	if b, err := r.cache.GetWithContext(ctx, url); err == nil && b != nil {
		return b, nil
	}

	resp, err := r.httpClient.GetWithContext(ctx, url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			r.log.Warn().Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = r.cache.SetWithContext(ctx, url, data, r.cache.DefaultExpiration()); err != nil {
		r.log.Warn().Err(err).Msg("Failed to cache response")
	}

	return data, nil
}

func (r *repository) getCoverFactory(ctx context.Context, userId int, id string, lang string, coverResp *MangaCoverResponse) CoverFactory {
	if len(coverResp.Data) == 0 {
		return defaultCoverFactory
	}

	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeCovers)
	defer span.End()

	covers, firstCover, lastCover := r.processCovers(ctx, id, lang, coverResp)
	return r.constructCoverFactory(ctx, userId, covers, firstCover, lastCover)
}

func (r *repository) processCovers(ctx context.Context, id string, lang string, coverResp *MangaCoverResponse) (map[string]string, string, string) {
	covers := make(map[string]string)
	var firstCover, lastCover, firstCoverLang, lastCoverLang string

	for _, cover := range coverResp.Data {
		// Don't download non-matching locale's again if a cover is already present
		if _, ok := covers[cover.Attributes.Volume]; ok && cover.Attributes.Locale != lang {
			continue
		}

		url := fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s.512.jpg", id, cover.Attributes.FileName)
		coverBytes, err := r.getCoverBytes(ctx, url)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, "", ""
			}
			r.log.Err(err).Str("fileName", cover.Attributes.FileName).Msg("Failed to get cover")
			continue
		}

		// Cover is too small for Kavita. Looks weird
		if !r.imageService.IsCover(coverBytes) {
			r.log.Debug().Str("id", cover.Id).Str("volume", cover.Attributes.Volume).
				Str("desc", cover.Attributes.Description).Msg("cover failed the ImageService.IsCover check. not using")
			continue
		}

		_cover := url

		if cover.Attributes.Locale == lang {
			covers[cover.Attributes.Volume] = _cover
			if firstCoverLang == "" {
				firstCoverLang = _cover
			}
			lastCoverLang = _cover
		} else if _, ok := covers[cover.Attributes.Volume]; !ok {
			covers[cover.Attributes.Volume] = _cover
		}

		if firstCover == "" {
			firstCover = _cover
		}
		lastCover = _cover
	}

	if firstCoverLang != "" {
		firstCover = firstCoverLang
	}
	if lastCoverLang != "" {
		lastCover = lastCoverLang
	}
	return covers, firstCover, lastCover
}

func (r *repository) constructCoverFactory(ctx context.Context, userId int, covers map[string]string, firstCover, lastCover string) CoverFactory {
	fallbackCover := func() string {
		p, err := r.unitOfWork.Preferences.GetPreferences(ctx, userId)
		if err != nil {
			return ""
		}

		switch p.CoverFallbackMethod {
		case models.CoverFallbackFirst:
			return firstCover
		case models.CoverFallbackLast:
			return lastCover
		case models.CoverFallbackNone:
			return ""
		}
		return firstCover
	}()

	return func(volume string) (string, bool) {
		cover, ok := covers[volume]
		if !ok && fallbackCover != "" {
			return fallbackCover, true
		}
		if !ok {
			return "", false
		}

		return cover, true
	}
}
