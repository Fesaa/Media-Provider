package webtoon

import (
	"errors"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
)

func loadImages(chapter Chapter) ([]string, error) {
	doc, err := wrapInDoc(chapter.Url)
	if err != nil {
		return nil, err
	}

	rawUrls := doc.Find("#_imageList img").Map(func(_ int, s *goquery.Selection) string {
		return s.AttrOr("data-url", "")
	})

	filteredUrls := utils.Filter(rawUrls, func(s string) bool {
		return s != ""
	})

	if len(filteredUrls) != len(rawUrls) {
		return nil, errors.New("not all img had a source")
	}

	return filteredUrls, nil
}
