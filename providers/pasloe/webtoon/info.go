package webtoon

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
)

func constructSeriesInfo(id string, httpClient *http.Client) (*Series, error) {
	seriesStartUrl := fmt.Sprintf(EPISODE_LIST, id)
	doc, err := wrapInDoc(seriesStartUrl, httpClient)
	if err != nil {
		return nil, err
	}

	series := &Series{}
	info := doc.Find(".detail_header .info")
	series.Genre = info.Find(".genre").Text()
	series.Name = info.Find(".subj").Text()
	series.Author = strings.TrimSpace(info.Find(".author_area").Children().Remove().End().Text())

	detail := doc.Find(".detail")
	series.Description = detail.Find(".summary").Text()
	series.Completed = strings.Contains(detail.Find(".day_info").Text(), "COMPLETED")
	series.Chapters = append(series.Chapters, extractChapters(doc)...)

	pages := utils.Filter(goquery.Map(doc.Find(".paginate a"), href), notEmpty)
	for index := 1; len(pages) > index; index++ {
		doc, err = wrapInDoc(DOMAIN+pages[index], httpClient)
		if err != nil {
			return nil, err
		}

		if index == len(pages)-1 && len(pages) > 10 {
			index = 1
		}

		series.Chapters = append(series.Chapters, extractChapters(doc)...)
		pages = utils.Filter(goquery.Map(doc.Find(".paginate a"), href), notEmpty)
		// Sleep a bit between these requests, to not spam them if the pages are a too high amount
		// The time is small enough to not matter, downloading the images will always take longer.
		time.Sleep(500 * time.Millisecond)
	}

	return series, nil
}

func extractChapters(doc *goquery.Document) (chapters []Chapter) {
	return goquery.Map(doc.Find("#_listUl li a"), func(_ int, s *goquery.Selection) (chapter Chapter) {
		chapter.Url = s.AttrOr("href", "")
		chapter.ImageUrl = s.Find("span img").AttrOr("src", "")
		chapter.Title = s.Find(".subj span").Text()
		chapter.Date = s.Find(".date").Text()
		chapter.Number = func() string {
			num := s.Find(".tx").Text()
			if len(num) > 0 && num[0] == '#' {
				return num[1:]
			}
			return num
		}()
		return chapter
	})
}

func notEmpty(s string) bool {
	return s != ""
}

func href(_ int, selection *goquery.Selection) string {
	return selection.AttrOr("href", "")
}
