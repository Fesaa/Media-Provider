package webtoon

import (
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"time"
)

func constructSeriesInfo(id string) (*Series, error) {
	seriesStartUrl := BASE_URL + id
	doc, err := wrapInDoc(seriesStartUrl)
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

	pages := utils.MaybeMap(goquery.Map(doc.Find(".paginate a"), href), notEmpty)
	for index := 1; len(pages) > index; index++ {
		doc, err = wrapInDoc(DOMAIN + pages[index])
		if err != nil {
			return nil, err
		}

		series.Chapters = append(series.Chapters, extractChapters(doc)...)
		pages = utils.MaybeMap(goquery.Map(doc.Find(".paginate a"), href), notEmpty)
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

func notEmpty(s string) (string, bool) {
	return s, s != ""
}

func href(_ int, selection *goquery.Selection) string {
	return selection.AttrOr("href", "")
}
