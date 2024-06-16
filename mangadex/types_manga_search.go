package mangadex

import "fmt"

type MangaSearchResponse MangaDexResponse[[]MangaSearchData]
type GetMangaResponse MangaDexResponse[MangaSearchData]

type MangaSearchData struct {
	Id            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    MangaAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`
}

func (a *MangaSearchData) RefURL() string {
	return fmt.Sprintf("https://mangadex.org/title/%s/", a.Id)
}

type MangaAttributes struct {
	Title            map[string]string `json:"title"`
	AltTitles        map[string]string `json:"alt_titles"`
	Description      map[string]string `json:"description"`
	OriginalLanguage string            `json:"originalLanguage"`
	LastVolume       string            `json:"lastVolume"`
	LastChapter      string            `json:"lastChapter"`
	Status           string            `json:"status"`
	Year             int               `json:"year"`
	ContentRating    string            `json:"contentRating"`
}

func (a *MangaAttributes) EnTitle() string {
	enAltTitle, ok := a.AltTitles["en"]
	if ok {
		return enAltTitle
	}

	enTitle, ok := a.Title["en"]
	if ok {
		return enTitle
	}
	return ""
}

func (a *MangaAttributes) EnDescription() string {
	enDescription, ok := a.Description["en"]
	if ok {
		return enDescription
	}
	return ""
}
