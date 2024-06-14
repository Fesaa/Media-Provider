package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"net/url"
)

const URL = "https://api.mangadex.org"

func addRange(u string, param string, r []string) string {
	for _, v := range r {
		u += fmt.Sprintf("&%s[]=%s", param, url.QueryEscape(v))
	}
	return u
}

func searchMangaURL(s SearchOptions) (string, error) {
	includedTagIds, err := mapTags(s.IncludedTags, s.SkipNotFoundTags)
	if err != nil {
		return "", err
	}
	excludedTagIds, err := mapTags(s.ExcludedTags, s.SkipNotFoundTags)
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

	base = addRange(base, "status", utils.Map(s.Status, func(t MangaStatus) string {
		return string(t)
	}))
	base = addRange(base, "contentRating", utils.Map(s.ContentRating, func(t ContentRating) string {
		return string(t)
	}))
	base += "&availableTranslatedLanguage[]=en"
	return base, nil
}

func chapterURL(id string) string {
	return fmt.Sprintf("%s/manga/%s/feed?order[volume]=desc&order[chapter]=desc", URL, id)
}

func chapterImageUrl(id string) string {
	return fmt.Sprintf("%s/at-home/server/%s", URL, id)
}

func getMangaURL(id string) string {
	return fmt.Sprintf("%s/manga/%s", URL, id)
}
