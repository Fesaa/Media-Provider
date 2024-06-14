package mangadex

type ChapterSearchResponse MangaDexResponse[ChapterSearchData]

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
