package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/go-metroninfo"
)

var linkConverter map[string]func(string) string

func init() {
	linkConverter = map[string]func(string) string{}

	linkConverter["al"] = func(s string) string {
		return fmt.Sprintf("https://anilist.co/manga/%s", s)
	}
	linkConverter["ap"] = func(s string) string {
		return fmt.Sprintf("https://www.anime-planet.com/manga/%s", s)
	}
	linkConverter["bw"] = func(s string) string {
		return fmt.Sprintf("https://bookwalker.jp/%s", s)
	}
	linkConverter["mu"] = func(s string) string {
		return fmt.Sprintf("https://www.mangaupdates.com/series.html?id=%s", s)
	}
	linkConverter["nu"] = func(s string) string {
		return fmt.Sprintf("https://www.novelupdates.com/series/%s", s)
	}
	linkConverter["kt"] = func(s string) string {
		return fmt.Sprintf("https://kitsu.io/api/edge/manga/%s", s)
	}
	linkConverter["amz"] = func(s string) string {
		return s
	}
	linkConverter["ebj"] = func(s string) string {
		return s
	}
	linkConverter["mal"] = func(s string) string {
		return fmt.Sprintf("https://myanimelist.net/manga/%s", s)
	}
	linkConverter["cdj"] = func(s string) string {
		return s
	}
	linkConverter["raw"] = func(s string) string {
		return s
	}
	linkConverter["engtl"] = func(s string) string {
		return s
	}
}

type MangaSearchResponse Response[[]MangaSearchData]
type GetMangaResponse Response[MangaSearchData]

type MangaSearchData struct {
	Id            string          `json:"id"`
	Type          string          `json:"type"`
	Attributes    MangaAttributes `json:"attributes"`
	Relationships []Relationship  `json:"relationships"`

	Language string `json:"-"`
}

func (a *MangaSearchData) GetId() string {
	return a.Id
}

func (a *MangaSearchData) GetTitle() string {
	return a.Attributes.LangTitle(a.Language)
}

func (a *MangaSearchData) AllChapters() []ChapterSearchData {
	panic("mangadex does not support getting chapters from the series")
}

func (a *MangaSearchData) RefUrl() string {
	return fmt.Sprintf("https://mangadex.org/title/%s/", a.Id)
}

func (a *MangaSearchData) CoverURL() string {
	cover := utils.Find(a.Relationships, func(r Relationship) bool {
		return r.Type == "cover_art"
	})
	if cover == nil {
		return ""
	}

	if fileName, ok := cover.Attributes["fileName"].(string); ok {
		// Link to the proxy endpoint of the api
		return fmt.Sprintf("proxy/mangadex/covers/%s/%s.256.jpg", a.Id, fileName)
	}

	return ""
}

func (a *MangaSearchData) Authors() []string {
	return utils.MaybeMap(a.Relationships, func(t Relationship) (string, bool) {
		if t.Type != "author" {
			return "", false
		}

		if name, ok := t.Attributes["name"].(string); ok {
			return name, true
		}
		return "", false
	})
}

func (a *MangaSearchData) Artists() []string {
	return utils.MaybeMap(a.Relationships, func(t Relationship) (string, bool) {
		if t.Type != "artist" {
			return "", false
		}

		if name, ok := t.Attributes["name"].(string); ok {
			return name, true
		}
		return "", false
	})
}

func (a *MangaSearchData) ScanlationGroup() []string {
	return utils.MaybeMap(a.Relationships, func(t Relationship) (string, bool) {
		if t.Type != "scanlation_group" {
			return "", false
		}

		if name, ok := t.Attributes["name"].(string); ok {
			return name, true
		}
		return "", false
	})
}

type MangaAttributes struct {
	Title            map[string]string   `json:"title"`
	AltTitles        []map[string]string `json:"altTitles"`
	Description      map[string]string   `json:"description"`
	IsLocked         bool                `json:"isLocked"`
	Links            map[string]string   `json:"links"`
	OriginalLanguage string              `json:"originalLanguage"`
	LastVolume       string              `json:"lastVolume"`
	LastChapter      string              `json:"lastChapter"`
	Status           MangaStatus         `json:"status"`
	Year             int                 `json:"year"`
	ContentRating    ContentRating       `json:"contentRating"`
	Tags             []TagData           `json:"tags"`
}

// LangTitle returns the best guess English title for the manga
// In order of tries; Title(language), AltTitles(language), Title(any), Fixed string with non 0 length
func (a *MangaAttributes) LangTitle(language string) string {
	// Note: for some reason the en title may still be in Japanese, don't really have a way of checking if it is
	// as the Japanese title is in the latin alphabet. We'll just have to be fine with it, as the alternative titles
	// are just plain weird from time to time
	langTitle, ok := a.Title[language]
	if ok && langTitle != "" {
		return langTitle
	}

	var langAltTitle string

	for _, altTitle := range a.AltTitles {
		if langAltTitle, ok = altTitle[language]; ok {
			break
		}
	}

	if langAltTitle != "" {
		return langAltTitle
	}

	// fallback to first title, any language
	for _, value := range a.Title {
		if value != "" {
			return value
		}
	}

	// Last resort fallback title, this should never happen- mangadex should have at least one title
	return "Media-Provider-Fallback-title"
}

func (a *MangaAttributes) LangAltTitles(language string) []string {
	var enAltTitles []string
	for _, altTitle := range a.AltTitles {
		for key, value := range altTitle {
			if key == language {
				enAltTitles = append(enAltTitles, value)
				break
			}
		}
	}
	return enAltTitles
}

func (a *MangaAttributes) LangDescription(language string) string {
	enDescription, ok := a.Description[language]
	if ok {
		return enDescription
	}
	return ""
}

func (a *MangaSearchData) FormattedLinks() []string {
	var out []string
	for key, link := range a.Attributes.Links {
		if conv, ok := linkConverter[key]; ok {
			out = append(out, conv(link))
		}
	}
	out = append(out, a.RefUrl())
	return out
}

type MangaStatus string

const (
	StatusOngoing   MangaStatus = "ongoing"
	StatusCompleted MangaStatus = "completed"
	StatusHiatus    MangaStatus = "hiatus"
	StatusCancelled MangaStatus = "cancelled"
)

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

func (c ContentRating) MetronInfoAgeRating() metroninfo.AgeRating {
	switch c {
	case ContentRatingSafe:
		return metroninfo.AgeRatingEveryone
	case ContentRatingSuggestive:
		return metroninfo.AgeRatingMature
	case ContentRatingErotica:
		return metroninfo.AgeRatingExplicit
	case ContentRatingPornographic:
		return metroninfo.AgeRatingAdult
	default:
		return metroninfo.AgeRatingUnknown
	}
}
