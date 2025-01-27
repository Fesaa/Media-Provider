package mangadex

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"io"
	"net/http"
)

type Repository interface {
	GetManga(id string) (*GetMangaResponse, error)
	SearchManga(options SearchOptions) (*MangaSearchResponse, error)
	GetChapters(id string, offset ...int) (*ChapterSearchResponse, error)
	GetChapterImages(id string) (*ChapterImageSearchResponse, error)
	GetCoverImages(id string, offset ...int) (*MangaCoverResponse, error)
}

type repository struct {
	httpClient *http.Client
	log        zerolog.Logger
	tags       *utils.SafeMap[string, string]
}

func NewRepository(httpClient *http.Client, log zerolog.Logger) Repository {
	r := &repository{
		httpClient: httpClient,
		log:        log,
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

func (r *repository) GetManga(id string) (*GetMangaResponse, error) {
	url := getMangaURL(id)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetManga")
	var getMangaResponse GetMangaResponse
	err := do(r.httpClient, url, &getMangaResponse)
	if err != nil {
		return nil, err
	}
	return &getMangaResponse, nil
}

func (r *repository) SearchManga(options SearchOptions) (*MangaSearchResponse, error) {
	url, err := r.searchMangaURL(options)
	if err != nil {
		return nil, err
	}

	var searchResponse MangaSearchResponse
	err = do(r.httpClient, url, &searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (r *repository) GetChapters(id string, offset ...int) (*ChapterSearchResponse, error) {
	url := chapterURL(id, offset...)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetChapters")
	var searchResponse ChapterSearchResponse
	err := do(r.httpClient, url, &searchResponse)
	if err != nil {
		return nil, err
	}

	if searchResponse.Total > searchResponse.Limit+searchResponse.Offset {
		extra, err := r.GetChapters(id, searchResponse.Limit+searchResponse.Offset)
		if err != nil {
			return nil, err
		}
		searchResponse.Data = append(searchResponse.Data, extra.Data...)
		return &searchResponse, nil
	}

	return &searchResponse, nil
}

func (r *repository) GetChapterImages(id string) (*ChapterImageSearchResponse, error) {
	url := chapterImageUrl(id)
	r.log.Trace().Str("id", id).Str("url", url).Msg("GetChapterImages")
	var searchResponse ChapterImageSearchResponse
	err := do(r.httpClient, url, &searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func (r *repository) GetCoverImages(id string, offset ...int) (*MangaCoverResponse, error) {
	url := getCoverURL(id, offset...)
	r.log.Trace().Str("id", id).Str("url", url).Str("offset", fmt.Sprintf("%#v", offset)).Msg("GetCoverImages")
	var searchResponse MangaCoverResponse
	err := do(r.httpClient, url, &searchResponse)
	if err != nil {
		return nil, err
	}

	if searchResponse.Total > searchResponse.Limit+searchResponse.Offset {
		extra, err := r.GetCoverImages(id, searchResponse.Limit+searchResponse.Offset)
		if err != nil {
			return nil, err
		}
		searchResponse.Data = append(searchResponse.Data, extra.Data...)
		return &searchResponse, nil
	}

	return &searchResponse, nil
}

func do[T any](httpClient *http.Client, url string, out *T) error {
	resp, err := httpClient.Get(url)
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
