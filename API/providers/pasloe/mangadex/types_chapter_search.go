package mangadex

import (
	"strconv"
)

type ChapterSearchResponse Response[[]ChapterSearchData]

type ChapterSearchData struct {
	Id            string            `json:"id"`
	Type          string            `json:"type"`
	Attributes    ChapterAttributes `json:"attributes"`
	Relationships []Relationship    `json:"relationships"`
}

func (chapter ChapterSearchData) GetChapter() string {
	return chapter.Attributes.Chapter
}

func (chapter ChapterSearchData) GetVolume() string {
	return chapter.Attributes.Volume
}

func (chapter ChapterSearchData) GetTitle() string {
	return chapter.Attributes.Title
}

func (chapter ChapterSearchData) GetId() string {
	return chapter.Id
}

func (chapter ChapterSearchData) Volume() float64 {
	if chapter.Attributes.Volume == "" {
		return -1
	}
	if vol, err := strconv.ParseFloat(chapter.Attributes.Volume, 64); err == nil {
		return vol
	}
	return -1
}

func (chapter ChapterSearchData) Chapter() float64 {
	if chapter.Attributes.Chapter == "" {
		return -1
	}

	if v, err := strconv.ParseFloat(chapter.Attributes.Chapter, 64); err == nil {
		return v
	}
	return -1
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
