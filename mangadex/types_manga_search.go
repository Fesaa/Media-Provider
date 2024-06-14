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
	Status           MangaStatus       `json:"status"`
	Year             int               `json:"year"`
	ContentRating    ContentRating     `json:"contentRating"`
}

func (a *MangaAttributes) EnTitle() string {
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

type PublicationDemographic string
type MangaStatus string
type ContentRating string

const (
	ONGOING   MangaStatus = "ongoing"
	COMPLETED MangaStatus = "completed"
	HIATUS    MangaStatus = "hiatus"
	CANCELLED MangaStatus = "cancelled"

	SHOUNEN PublicationDemographic = "shounen"
	SHOUJO  PublicationDemographic = "shounjo"
	SEINEN  PublicationDemographic = "seinen"
	JOSEIN  PublicationDemographic = "josein"

	SAFE       ContentRating = "safe"
	SUGGESTIVE ContentRating = "suggestive"
	EROTICA    ContentRating = "erotica"
)
