package dynasty_scans

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strings"
)

const (
	DOMAIN = "https://dynasty-scans.com"

	SEARCH = DOMAIN + "/search?q=%s&classes[]=Series"
)

func SearchSeries(opt SearchOptions) ([]SearchData, error) {
	doc, err := wrapInDoc(searchUrl(opt.Query))
	if err != nil {
		return nil, err
	}

	series := doc.Find(".chapter-list dd")
	return goquery.Map(series, selectionToSearchData), nil
}

func selectionToSearchData(_ int, sel *goquery.Selection) SearchData {
	sd := SearchData{}

	nameElement := sel.Find(".name").First()
	sd.Title = nameElement.Text()
	sd.Id = func() string {
		ref := nameElement.AttrOr("href", "")
		if ref == "" {
			return ref
		}

		if strings.HasPrefix(ref, "/series/") {
			return strings.TrimPrefix(ref, "/series/")
		}
		return ref
	}()

	sd.Tags = sel.Find(".tags a").Map(func(_ int, s *goquery.Selection) string {
		return s.Text()
	})

	sd.Authors = sel.Find("a").Map(func(_ int, s *goquery.Selection) string {
		return s.Text()
	})

	return sd
}

func searchUrl(keyword string) string {
	return fmt.Sprintf(SEARCH, url.QueryEscape(keyword))
}
