package mangadex

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/http/wisewolf"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"net/http"
)

var tags = utils.NewSafeMap[string, string]()

func mapTags(in []string, skip bool) ([]string, error) {
	mappedTags := make([]string, 0)
	for _, tag := range in {
		id, ok := tags.Get(tag)
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

func GetManga(id string) (*GetMangaResponse, error) {
	url := getMangaURL(id)
	log.Trace("getting manga info", "mangaId", id, "url", url)
	var getMangaResponse GetMangaResponse
	err := do(url, &getMangaResponse)
	if err != nil {
		return nil, err
	}
	return &getMangaResponse, nil
}

func SearchManga(options SearchOptions) (*MangaSearchResponse, error) {
	url, err := searchMangaURL(options)
	if err != nil {
		return nil, err
	}

	var searchResponse MangaSearchResponse
	err = do(url, &searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func GetChapters(id string, offset ...int) (*ChapterSearchResponse, error) {
	url := chapterURL(id, offset...)
	log.Trace("getting chapters", "mangaId", id, "url", url)
	var searchResponse ChapterSearchResponse
	err := do(url, &searchResponse)
	if err != nil {
		return nil, err
	}

	if searchResponse.Total > searchResponse.Limit+searchResponse.Offset {
		extra, err := GetChapters(id, searchResponse.Limit+searchResponse.Offset)
		if err != nil {
			return nil, err
		}
		searchResponse.Data = append(searchResponse.Data, extra.Data...)
		return &searchResponse, nil
	}

	return &searchResponse, nil
}

func GetChapterImages(id string) (*ChapterImageSearchResponse, error) {
	url := chapterImageUrl(id)
	log.Trace("getting chapter images", "mangaId", id, "url", url)
	var searchResponse ChapterImageSearchResponse
	err := do(url, &searchResponse)
	if err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func GetCoverImages(id string, offset ...int) (*MangaCoverResponse, error) {
	url := getCoverURL(id, offset...)
	log.Trace("getting cover images", "mangaId", id, "offset", fmt.Sprintf("%#v", offset), "url", url)
	var searchResponse MangaCoverResponse
	err := do(url, &searchResponse)
	if err != nil {
		return nil, err
	}

	if searchResponse.Total > searchResponse.Limit+searchResponse.Offset {
		extra, err := GetCoverImages(id, searchResponse.Limit+searchResponse.Offset)
		if err != nil {
			return nil, err
		}
		searchResponse.Data = append(searchResponse.Data, extra.Data...)
		return &searchResponse, nil
	}

	return &searchResponse, nil
}

func do[T any](url string, out *T) error {
	resp, err := wisewolf.Client.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Warn("failed to close body", "error", err)
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(data, out); err != nil {
		return err
	}
	return nil
}
