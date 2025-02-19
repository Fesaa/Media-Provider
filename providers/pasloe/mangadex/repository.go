package mangadex

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"io"
	"net/http"
)

type Repository interface {
	GetManga(ctx context.Context, id string) (*GetMangaResponse, error)
	SearchManga(ctx context.Context, options SearchOptions) (*MangaSearchResponse, error)
	GetChapters(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error)
	GetChapterImages(ctx context.Context, id string) (*ChapterImageSearchResponse, error)
	GetCoverImages(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error)
}

type repository struct {
	httpClient *http.Client
	log        zerolog.Logger
	tags       utils.SafeMap[string, string]
}

type repositoryParams struct {
	dig.In

	HttpClient *http.Client `name:"http-retry"`
}

func NewRepository(params repositoryParams, log zerolog.Logger) Repository {
	r := &repository{
		httpClient: params.HttpClient,
		log:        log.With().Str("handler", "mangadex-repository").Logger(),
		tags:       utils.NewSafeMap[string, string](),
	}
	if err := r.loadTags(); err != nil {
		r.log.Error().Err(err).Msg("failed to load tags, some features may not work")
	} else {
		r.log.Debug().Int("size", r.tags.Len()).Msg("loaded tags")
	}
	return r
}

func (r *repository) loadTags() error {
	tagURL := URL + "/manga/tag"

	resp, err := r.httpClient.Get(tagURL)
	if err != nil {
		return fmt.Errorf("loadTags Get: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("loadTags status: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("loadTags readAll: %w", err)
	}

	var tagResponse TagResponse
	err = json.Unmarshal(body, &tagResponse)
	if err != nil {
		return fmt.Errorf("loadTags unmarshal: %w", err)
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
	err := do(ctx, r.httpClient, url, &getMangaResponse)
	if err != nil {
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
	err = do(ctx, r.httpClient, url, &searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (r *repository) GetChapters(ctx context.Context, id string, offset ...int) (*ChapterSearchResponse, error) {
	url := chapterURL(id, offset...)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetChapters")
	var searchResponse ChapterSearchResponse
	err := do(ctx, r.httpClient, url, &searchResponse)
	if err != nil {
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
	err := do(ctx, r.httpClient, url, &searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (r *repository) GetCoverImages(ctx context.Context, id string, offset ...int) (*MangaCoverResponse, error) {
	url := getCoverURL(id, offset...)
	r.log.Trace().Str("id", id).Str("url", url).Str("offset", fmt.Sprintf("%#v", offset)).Msg("GetCoverImages")
	var searchResponse MangaCoverResponse
	err := do(ctx, r.httpClient, url, &searchResponse)
	if err != nil {
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

func do[T any](ctx context.Context, httpClient *http.Client, url string, out *T) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}
