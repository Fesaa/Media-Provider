package mangadex

import (
	"fmt"
	"net/url"
)

const URL = "https://api.mangadex.org"

func addRange(u string, param string, r []string) string {
	for _, v := range r {
		u += fmt.Sprintf("&%s[]=%s", param, url.QueryEscape(v))
	}
	return u
}

func (r *repository) searchMangaURL(s SearchOptions) (string, error) {
	includedTagIds, err := r.mapTags(s.IncludedTags, s.SkipNotFoundTags)
	if err != nil {
		return "", err
	}
	excludedTagIds, err := r.mapTags(s.ExcludedTags, s.SkipNotFoundTags)
	if err != nil {
		return "", err
	}

	base := URL + "/manga?"
	base += "title=" + url.QueryEscape(s.Query)

	if len(includedTagIds) > 0 {
		base = addRange(base, "includedTags", includedTagIds)
		base += "&includedTagsMode=OR"
	}

	if len(excludedTagIds) > 0 {
		base = addRange(base, "excludedTags", excludedTagIds)
		base += "&excludedTagsMode=OR"
	}

	base = addRange(base, "status", s.Status)
	base = addRange(base, "contentRating", s.ContentRating)
	base += "&includes[]=cover_art"
	base += "&includes[]=author"
	base += "&includes[]=artist"
	base += "&availableTranslatedLanguage[]=en"
	base += "&limit=20"
	return base, nil
}

func chapterURL(id string, offset ...int) string {
	contentRatingSuffix := "&contentRating[]=pornographic&contentRating[]=erotica&contentRating[]=suggestive&contentRating[]=safe"
	if len(offset) > 0 {
		return fmt.Sprintf("%s/manga/%s/feed?order[volume]=desc&order[chapter]=desc&offset=%d%s", URL, id, offset[0], contentRatingSuffix)
	}

	return fmt.Sprintf("%s/manga/%s/feed?order[volume]=desc&order[chapter]=desc%s", URL, id, contentRatingSuffix)
}

func chapterImageUrl(id string) string {
	return fmt.Sprintf("%s/at-home/server/%s", URL, id)
}

func getMangaURL(id string) string {
	suffix := "?includes[]=cover_art"
	suffix += "&includes[]=author"
	suffix += "&includes[]=artist"
	return fmt.Sprintf("%s/manga/%s%s", URL, id, suffix)
}

func getCoverURL(id string, offset ...int) string {
	if len(offset) > 0 {
		return fmt.Sprintf("%s/cover/?limit=20&manga[]=%s&offset=%d", URL, id, offset[0])
	}

	return fmt.Sprintf("%s/cover/?limit=20&manga[]=%s", URL, id)
}
