package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

const timeLayout = "2006-01-02T15:04:05Z07:00"

type Repository interface {
	SearchManga(ctx context.Context, options SearchOptions) (*MangaSearchResponse, error)
	SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error)
	ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error)
}

type repository struct {
	httpClient   *menou.Client
	cache        services.CacheService
	imageService services.ImageService
	unitOfWork   *db.UnitOfWork
	log          zerolog.Logger
	tags         utils.SafeMap[string, string]
}

type repositoryParams struct {
	dig.In

	HttpClient   *menou.Client `name:"http-retry"`
	Cache        services.CacheService
	ImageService services.ImageService
	UnitOfWork   *db.UnitOfWork
	Ctx          context.Context
}

func NewRepository(params repositoryParams, log zerolog.Logger) Repository {
	ctx, span := tracing.TracerServices.Start(params.Ctx, tracing.SetupRepository,
		trace.WithAttributes(attribute.String("repository.name", "Mangadex")))
	defer span.End()

	r := &repository{
		httpClient:   params.HttpClient,
		cache:        params.Cache,
		imageService: params.ImageService,
		unitOfWork:   params.UnitOfWork,
		log:          log.With().Str("handler", "mangadex-repository").Logger(),
		tags:         utils.NewSafeMap[string, string](),
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := r.loadTags(ctx); err != nil {
		r.log.Error().Err(err).Msg("failed to load tags, some features may not work")
	} else {
		r.log.Debug().Int("size", r.tags.Len()).Msg("loaded tags")
	}

	return r
}

func (r *repository) SearchManga(ctx context.Context, options SearchOptions) (*MangaSearchResponse, error) {
	url, err := r.searchMangaURL(options)
	if err != nil {
		return nil, err
	}

	var searchResponse MangaSearchResponse
	if err = r.do(ctx, url, &searchResponse); err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (r *repository) SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error) {
	m, err := r.GetManga(ctx, id)
	if err != nil {
		return publication.Series{}, err
	}

	chapters, err := r.GetChapters(ctx, id)
	if err != nil {
		return publication.Series{}, err
	}

	lang := req.GetStringOrDefault(LanguageKey, "en")
	altTitle := utils.OrDefault(m.Data.Attributes.LangAltTitles(lang), "")
	chapters = r.FilterChapters(lang, req, chapters)

	tags := utils.MaybeMap(m.Data.Attributes.Tags, func(t TagData) (publication.Tag, bool) {
		tag, ok := t.Attributes.Name[lang]
		if !ok {
			return publication.Tag{}, false
		}

		return publication.Tag{
			Value:      tag,
			Identifier: t.Id,
			IsGenre:    t.Attributes.Group == "genre",
		}, true
	})

	return publication.Series{
		Id:                id,
		Title:             m.Data.Attributes.LangTitle(lang),
		AltTitle:          altTitle,
		Description:       m.Data.Attributes.LangDescription(lang),
		RefUrl:            m.Data.RefUrl(),
		Status:            toPublicationStatus(m.Data.Attributes.Status),
		TranslationStatus: utils.Settable[publication.Status]{},
		HighestVolume:     utils.NewSettableFromErr(strconv.ParseFloat(m.Data.Attributes.LastVolume, 64)),
		HighestChapter:    utils.NewSettableFromErr(strconv.ParseFloat(m.Data.Attributes.LastChapter, 64)),
		Year:              m.Data.Attributes.Year,
		OriginalLanguage:  m.Data.Attributes.OriginalLanguage,
		ContentRating:     m.Data.Attributes.ContentRating.ComicInfoAgeRating(),
		Tags:              tags,
		People:            m.Data.People(),
		Links:             m.Data.FormattedLinks(),
		Chapters:          utils.Map(chapters.Data, r.mapChapter),
	}, nil
}

func (r *repository) mapChapter(data ChapterSearchData) publication.Chapter {

	var releaseDate *time.Time
	if data.Attributes.PublishedAt != "" {
		t, err := time.Parse(timeLayout, data.Attributes.PublishedAt)
		if err != nil {
			r.log.Warn().Err(err).Msg("failed to parse published time")
			releaseDate = &time.Time{}
		} else {
			releaseDate = &t
		}
	}

	translator := utils.MaybeMap(data.Relationships, func(r Relationship) (string, bool) {
		if r.Type != "scanlation_group" && r.Type != "user" {
			return "", false
		}

		return r.Id, true
	})

	return publication.Chapter{
		Id:          data.Id,
		Title:       data.Attributes.Title,
		Volume:      data.Attributes.Volume,
		Chapter:     data.Attributes.Chapter,
		CoverUrl:    "",
		ReleaseDate: releaseDate,
		Translator:  translator,
		Tags:        nil,
		People:      nil,
	}
}

func (r *repository) FilterChapters(lang string, req payload.DownloadRequest, c *ChapterSearchResponse) *ChapterSearchResponse {
	scanlation := req.GetStringOrDefault(publication.ScanlationGroupKey, "")
	allowNonMatching := req.GetBool(AllowNonMatchingScanlationGroupKey, true)

	chaptersMap := utils.GroupBy(c.Data, func(v ChapterSearchData) string {
		if v.Attributes.Chapter == "" { // group OneShots
			return ""
		}
		return v.Attributes.Chapter + " - " + v.Attributes.Volume
	})

	newData := make([]ChapterSearchData, 0)
	for marker, chapters := range chaptersMap {
		// OneShots are handled later
		if marker == "" {
			continue
		}

		chapter, ok := utils.FindOk(chapters, r.chapterSearchFunc(lang, scanlation, true))

		// Retry by skipping scanlation check
		if !ok && scanlation != "" && allowNonMatching {
			chapter, ok = utils.FindOk(chapters, r.chapterSearchFunc(lang, "", true))
		}

		if ok {
			newData = append(newData, chapter)
		}
	}

	if req.GetBool(publication.DownloadOneShotKey) {
		// OneShots do not have a chapter, so will be mapped under the empty string
		if chapters, ok := chaptersMap[""]; ok {
			newData = append(newData, utils.Filter(chapters, r.chapterSearchFunc(lang, scanlation, false))...)
		}
	}

	c.Data = newData
	return c
}

func (r *repository) chapterSearchFunc(lang, scanlation string, skipOneShot bool) func(ChapterSearchData) bool {
	return func(data ChapterSearchData) bool {
		if data.Attributes.TranslatedLanguage != lang {
			return false
		}
		// Skip over official publisher chapters, we cannot download these from mangadex
		if data.Attributes.ExternalUrl != "" {
			return false
		}

		if data.Attributes.Chapter == "" && skipOneShot {
			return false
		}

		if scanlation == "" {
			return true
		}

		return slices.ContainsFunc(data.Relationships, func(relationship Relationship) bool {
			if relationship.Type != "scanlation_group" && relationship.Type != "user" {
				return false
			}

			return relationship.Id == scanlation
		})
	}
}

func (r *repository) PreDownloadHook(p publication.Publication, ctx context.Context) error {
	r.log.Debug().Msg("loading covers for chapters")

	covers, err := r.GetCoverImages(ctx, p.Request().Id)
	if err != nil {
		return err
	}

	lang := p.Request().GetStringOrDefault(LanguageKey, "en")
	coverFactory := r.getCoverFactory(ctx, p.Request().OwnerId, p.Request().Id, lang, covers)

	p.UpdateSeriesInfo(func(series *publication.Series) {
		series.Chapters = utils.Map(series.Chapters, func(chapter publication.Chapter) publication.Chapter {
			if url, ok := coverFactory(chapter.Volume); ok {
				chapter.CoverUrl = url
			}
			return chapter
		})
	})

	return nil
}

func (r *repository) HttpGetHook(req *http.Request) error {
	req.Header.Add(fiber.HeaderOrigin, "https://mangadex.org")
	return nil
}

func (r *repository) ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error) {
	url := chapterImageUrl(chapter.Id)
	r.log.Trace().Str("id", chapter.Id).Str("url", url).Msg("GetChapterImages")
	var searchResponse ChapterImageSearchResponse
	if err := r.do(ctx, url, &searchResponse); err != nil {
		return nil, err
	}

	r.log.Debug().
		Str("id", chapter.Id).
		Str("baseUrl", searchResponse.BaseUrl).
		Msg("chapter images will be retrieved from a home server")
	return searchResponse.FullImageUrls(), nil
}

func (r *repository) loadTags(ctx context.Context) error {
	tagURL := URL + "/manga/tag"

	var tagResponse TagResponse
	if err := r.do(ctx, tagURL, &tagResponse, r.cache.DefaultExpiration()); err != nil {
		return err
	}

	for _, tag := range tagResponse.Data {
		enName, ok := tag.Attributes.Name["en"]
		if !ok {
			continue
		}
		r.tags.Set(enName, tag.Id)
	}
	return nil
}

func (r *repository) mapTags(in []string, skip bool) ([]string, error) {
	mappedTags := make([]string, 0)
	for _, tag := range in {
		id, ok := r.tags.Get(tag)
		if !ok {
			if skip {
				continue
			}
			return nil, fmt.Errorf("tag %s not found", tag)
		}
		mappedTags = append(mappedTags, id)
	}
	return mappedTags, nil
}

func (r *repository) GetManga(ctx context.Context, id string) (*GetMangaResponse, error) {
	url := getMangaURL(id)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetManga")
	var getMangaResponse GetMangaResponse
	if err := r.do(ctx, url, &getMangaResponse); err != nil {
		return nil, err
	}
	return &getMangaResponse, nil
}

func (r *repository) GetChapters(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
	url := chapterURL(id, offset...)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetChapters")
	var searchResponse ChapterSearchResponse
	if err := r.do(ctx, url, &searchResponse); err != nil {
		return nil, err
	}

	if searchResponse.Total > searchResponse.Limit+searchResponse.Offset {
		extra, err := r.GetChapters(ctx, id, searchResponse.Limit+searchResponse.Offset)
		if err != nil {
			return nil, err
		}
		searchResponse.Data = append(searchResponse.Data, extra.Data...)
		return &searchResponse, nil
	}

	return &searchResponse, nil
}

func (r *repository) GetCoverImages(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error) {
	url := getCoverURL(id, offset...)
	r.log.Trace().Str("id", id).Str("url", url).Str("offset", fmt.Sprintf("%#v", offset)).Msg("GetCoverImages")
	var searchResponse MangaCoverResponse
	if err := r.do(ctx, url, &searchResponse); err != nil {
		return nil, err
	}

	if searchResponse.Total > searchResponse.Limit+searchResponse.Offset {
		extra, err := r.GetCoverImages(ctx, id, searchResponse.Limit+searchResponse.Offset)
		if err != nil {
			return nil, err
		}
		searchResponse.Data = append(searchResponse.Data, extra.Data...)
		return &searchResponse, nil
	}

	return &searchResponse, nil
}

func (r *repository) do(ctx context.Context, url string, out any, exp ...time.Duration) error {
	ctx, span := tracing.TracerPasloe.Start(ctx, tracing.SpanPasloeCachedDownload)
	defer span.End()

	if v, err := r.cache.GetWithContext(ctx, url); err == nil && v != nil {
		if err = json.Unmarshal(v, out); err == nil {
			return nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, out); err != nil {
		return err
	}

	if err = r.cache.SetWithContext(ctx, url, data, utils.OrDefault(exp, time.Minute*5)); err != nil {
		r.log.Debug().Err(err).Str("key", url).Msg("failed to set cache for outgoing mangadex request")
	}
	return nil

}
