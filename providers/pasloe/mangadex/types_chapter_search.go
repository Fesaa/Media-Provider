package mangadex

type ChapterSearchResponse Response[[]ChapterSearchData]

func (c ChapterSearchResponse) FilterToLanguage(language string) ChapterSearchResponse {
	c2 := c
	newData := make([]ChapterSearchData, 0)

	lastChapter := "random stuff that will never match"
	lastVolume := "random stuff that will never match"
	for _, data := range c.Data {
		if data.Attributes.Volume == lastVolume && data.Attributes.Chapter == lastChapter {
			continue
		}
		if data.Attributes.TranslatedLanguage != language {
			continue
		}

		// Skip over official publisher chapters, we cannot download these from mangadex
		if data.Attributes.ExternalUrl != "" {
			continue
		}

		newData = append(newData, data)
		lastChapter = data.Attributes.Chapter
		lastVolume = data.Attributes.Volume
	}

	c2.Data = newData
	return c2
}

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
