package mangadex

type MangaSearchResponse MangaDexResponse[MangaSearchData]

type MangaSearchData struct {
	Id            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    MangaAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`
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
	SUGGESTIVE ContentRating = "suggested"
	EROTICA    ContentRating = "erotica"
)
