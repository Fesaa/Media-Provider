package mangadex

import (
	"fmt"

	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
)

type SearchOptions struct {
	Query                  string
	IncludedTags           []string
	IncludeTagsMode        string
	ExcludedTags           []string
	ExcludedTagsMode       string
	Status                 []string
	ContentRating          []string
	PublicationDemographic []string
	SkipNotFoundTags       bool
}

type Response[T any] struct {
	Result   string `json:"result"`
	Response string `json:"response"`
	Data     T      `json:"data"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Total    int    `json:"total"`
}

type Relationship struct {
	Id         string         `json:"id"`
	Type       string         `json:"type"`
	Related    string         `json:"related,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type ChapterImageSearchResponse struct {
	Result  string      `json:"result"`
	BaseUrl string      `json:"baseUrl"`
	Chapter ChapterInfo `json:"chapter"`
}

// FullImageUrls returns the urls for full quality
func (s *ChapterImageSearchResponse) FullImageUrls() []publication.DownloadUrl {
	urls := make([]publication.DownloadUrl, len(s.Chapter.Data))
	for i, image := range s.Chapter.Data {
		urls[i] = publication.DownloadUrl{
			Url: fmt.Sprintf("%s/data/%s/%s", s.BaseUrl, s.Chapter.Hash, image),
			// Mangadex is timing out on single chapter images. For these we'll get them from the fallback
			FallbackUrl: fmt.Sprintf("%s/data/%s/%s", "https://uploads.mangadex.org", s.Chapter.Hash, image),
		}
	}
	return urls
}

type ChapterInfo struct {
	Hash      string   `json:"hash"`
	Data      []string `json:"data"`
	DataSaver []string `json:"dataSaver"`
}

type ChapterSearchResponse Response[[]ChapterSearchData]

type ChapterSearchData struct {
	Id            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    ChapterAttributes `json:"attributes"`
	Relationships []Relationship    `json:"relationships"`
}

type ChapterAttributes struct {
	Volume             string `json:"volume"`
	Chapter            string `json:"chapter"`
	Title              string `json:"title"`
	TranslatedLanguage string `json:"translatedLanguage"`
	ExternalUrl        string `json:"externalUrl"`
	PublishedAt        string `json:"publishedAt"`
	ReadableAt         string `json:"readableAt"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	Pages              int    `json:"pages"`
	Version            int    `json:"version"`
}

type TagResponse Response[[]TagData]

type TagData struct {
	Id         string        `json:"id"`
	Type       string        `json:"type"`
	Attributes TagAttributes `json:"attributes"`
}

type TagAttributes struct {
	Name          map[string]string `json:"name"`
	Description   map[string]string `json:"description"`
	Group         string            `json:"group"`
	Version       int               `json:"version"`
	Relationships []Relationship    `json:"relationships"`
}
