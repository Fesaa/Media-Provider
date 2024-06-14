package mangadex

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
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

func GetManga(id string) (*GetMangaResponse, error, time.Duration) {
	url := getMangaURL(id)
	slog.Debug("Getting manga info", "id", id, "url", url)
	var getMangaResponse GetMangaResponse
	return do(url, &getMangaResponse)
}

func SearchManga(options SearchOptions) (*MangaSearchResponse, error) {
	url, err := searchMangaURL(options)
	if err != nil {
		return nil, err
	}
	slog.Debug("Searching Mangadex for Manga", "options", fmt.Sprintf("%#v", options), "url", url)
	var searchResponse MangaSearchResponse
	a, b, _ := do(url, &searchResponse)
	return a, b
}

func GetChapters(id string) (*ChapterSearchResponse, error, time.Duration) {
	url := chapterURL(id)
	slog.Debug("Getting chapters", "id", id, "url", url)
	var searchResponse ChapterSearchResponse
	return do(url, &searchResponse)
}

func GetChapterImages(id string) (*ChapterImageSearchResponse, error, time.Duration) {
	url := chapterImageUrl(id)
	slog.Debug("Getting chapter images", "id", id, "url", url)
	var searchResponse ChapterImageSearchResponse
	return do(url, &searchResponse)
}

func do[T any](url string, out *T) (*T, error, time.Duration) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err, 0
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		unixRateLimitEnd := resp.Header.Get("X-RateLimit-Retry-After")
		if unixRateLimitEnd == "" {
			return nil, fmt.Errorf("too many requests"), 1 * time.Minute
		}
		unix, err := strconv.ParseInt(unixRateLimitEnd, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("too many requests"), 1 * time.Minute
		}
		return nil, fmt.Errorf("too many requests"), time.Now().Sub(time.Unix(unix, 0))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status), 0
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err, 0
	}

	if err := json.Unmarshal(data, out); err != nil {
		return nil, err, 0
	}
	return out, nil, 0
}
