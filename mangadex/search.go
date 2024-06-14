package mangadex

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"log/slog"
	"net/http"
)

var tags = utils.NewSafeMap[string, string]()

func mapTags(in []string, skip bool) ([]string, error) {
	mappedTags := make([]string, len(in))
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

func SearchManga(options SearchOptions) (*MangaSearchResponse, error) {
	url, err := searchMangaURL(options)
	if err != nil {
		return nil, err
	}
	slog.Debug("Searching Mangadex for Manga", "options", fmt.Sprintf("%#v", options), "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResponse MangaSearchResponse
	if err := json.Unmarshal(data, &searchResponse); err != nil {
		return nil, err
	}

	return &searchResponse, nil
}

func GetChapters(id string) (*ChapterSearchResponse, error) {
	url := chapterURL(id)

	slog.Debug("Getting chapters", "id", id, "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResponse ChapterSearchResponse
	if err := json.Unmarshal(data, &searchResponse); err != nil {
		return nil, err
	}
	return &searchResponse, nil
}

func GetChapterImages(id string) (*ChapterImageSearchResponse, error) {
	url := chapterImageUrl(id)

	slog.Debug("Getting chapter images", "id", id, "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var searchResponse ChapterImageSearchResponse
	if err := json.Unmarshal(data, &searchResponse); err != nil {
		return nil, err
	}
	return &searchResponse, nil
}
