package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
)

type MangaSearchResponse Response[[]MangaSearchData]
type GetMangaResponse Response[MangaSearchData]

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
	Title            map[string]string   `json:"title"`
	AltTitles        []map[string]string `json:"altTitles"`
	Description      map[string]string   `json:"description"`
	OriginalLanguage string              `json:"originalLanguage"`
	LastVolume       string              `json:"lastVolume"`
	LastChapter      string              `json:"lastChapter"`
	Status           string              `json:"status"`
	Year             int                 `json:"year"`
	ContentRating    ContentRating       `json:"contentRating"`
	Tags             []TagData           `json:"tags"`
}

func (a *MangaAttributes) EnTitle() string {
	// Note: for some reason the en title may still be in Japanese, don't really have a way of checking if it is
	// as the Japanese title is in the latin alphabet. We'll just have to be fine with it, as the alternative titles
	// are just plain weird from time to time
	enTitle, ok := a.Title["en"]
	if ok {
		return enTitle
	}

	var enAltTitle string
	for _, altTitle := range a.AltTitles {
		for key, value := range altTitle {
			if key == "en" {
				enAltTitle = value
				break
			}
		}
	}

	if enAltTitle != "" {
		return enAltTitle
	}
	return ""
}

func (a *MangaAttributes) EnAltTitles() []string {
	var enAltTitles []string
	for _, altTitle := range a.AltTitles {
		for key, value := range altTitle {
			if key == "en" {
				enAltTitles = append(enAltTitles, value)
				break
			}
		}
	}
	return enAltTitles
}

func (a *MangaAttributes) EnDescription() string {
	enDescription, ok := a.Description["en"]
	if ok {
		return enDescription
	}
	return ""
}

type ContentRating string

const (
	ContentRatingSafe         ContentRating = "safe"
	ContentRatingSuggestive   ContentRating = "suggestive"
	ContentRatingErotica      ContentRating = "erotica"
	ContentRatingPornographic ContentRating = "pornographic"
)

func (c ContentRating) ComicInfoAgeRating() comicinfo.AgeRating {
	switch c {
	case ContentRatingSafe:
		return comicinfo.AgeRatingEveryone
	case ContentRatingSuggestive:
		return comicinfo.AgeRatingTeen
	case ContentRatingErotica:
		return comicinfo.AgeRatingMaturePlus17
	case ContentRatingPornographic:
		return comicinfo.AgeRatingAdultsOnlyPlus18
	default:
		return comicinfo.AgeRatingUnknown
	}
}
