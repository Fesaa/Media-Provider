package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type Repository interface {
	GetManga(ctx context.Context, id string) (*GetMangaResponse, error)
	SearchManga(ctx context.Context, options SearchOptions) (*MangaSearchResponse, error)
	GetChapters(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error)
	GetChapterImages(ctx context.Context, id string) (*ChapterImageSearchResponse, error)
	GetCoverImages(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error)
}

type repository struct {
	httpClient *menou.Client
	cache      services.CacheService
	log        zerolog.Logger
	tags       utils.SafeMap[string, string]
}

type repositoryParams struct {
	dig.In

	HttpClient *menou.Client `name:"http-retry"`
	Cache      services.CacheService
	Ctx        context.Context
}

func NewRepository(params repositoryParams, log zerolog.Logger) Repository {
	ctx, span := tracing.TracerServices.Start(params.Ctx, tracing.SetupRepository,
		trace.WithAttributes(attribute.String("repository.name", "Mangadex")))
	defer span.End()

	r := &repository{
		httpClient: params.HttpClient,
		cache:      params.Cache,
		log:        log.With().Str("handler", "mangadex-repository").Logger(),
		tags:       utils.NewSafeMap[string, string](),
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

func (r *repository) loadTags(ctx context.Context) error {
	tagURL := URL + "/manga/tag"

	var tagResponse TagResponse
	if err := r.do(ctx, tagURL, &tagResponse, 24*time.Hour); err != nil {
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

func (r *repository) GetChapterImages(ctx context.Context, id string) (*ChapterImageSearchResponse, error) {
	url := chapterImageUrl(id)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetChapterImages")
	var searchResponse ChapterImageSearchResponse
	if err := r.do(ctx, url, &searchResponse); err != nil {
		return nil, err
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
