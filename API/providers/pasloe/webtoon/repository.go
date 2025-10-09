package webtoon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const (
	Domain      = "https://www.webtoons.com"
	BaseUrl     = "https://www.webtoons.com/en/"
	SearchUrl   = BaseUrl + "search?keyword=%s"
	ImagePrefix = "https://webtoon-phinf.pstatic.net/"
	EpisodeList = Domain + "/episodeList?titleNo=%s"
)

var (
	rg = regexp.MustCompile("[^a-zA-Z0-9 ]+")

	ErrMissingSource = errors.New("not all img had a source")
)

type Repository interface {
	Search(ctx context.Context, options SearchOptions) ([]SearchData, error)
	ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error)
	SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error)
}

type repository struct {
	httpClient *menou.Client
	log        zerolog.Logger
}

func NewRepository(httpClient *menou.Client, log zerolog.Logger) Repository {
	return &repository{
		httpClient: httpClient,
		log:        log,
	}
}

func (r *repository) Search(ctx context.Context, options SearchOptions) ([]SearchData, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, searchUrl(options.Query), r.HttpGetHook)
	if err != nil {
		return nil, err
	}

	var results []SearchData
	results = append(results, goquery.Map(doc.Find(".card_lst li"), r.extractSeries)...)
	// results = append(results, goquery.Map(doc.Find(".challenge_lst ul li"), r.extractSeries)...) // Canvas
	return results, nil
}

func (r *repository) extractSeries(_ int, s *goquery.Selection) SearchData {
	id := s.Find("a").First().AttrOr("data-title-no", "")
	rating := s.Find("a").First().AttrOr("data-title-unsuitable-for-children", "false")

	return SearchData{
		Id:              id,
		Name:            s.Find(".subj").Text(),
		ReadCount:       s.Find("em.grade_num").Text(),
		ThumbnailMobile: s.Find("img").AttrOr("src", ""),
		AuthorNameList:  utils.Map(strings.Split(s.Find(".author").Text(), "/"), strings.TrimSpace),
		Genre:           s.Find(".genre").Text(),
		Rating:          rating != "false",
	}
}

func (r *repository) ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, chapter.Url, r.HttpGetHook)
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
		return nil, ErrMissingSource
	}

	return filteredUrls, nil
}

func (r *repository) SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error) {
	seriesStartUrl := fmt.Sprintf(EpisodeList, id)
	doc, err := r.httpClient.WrapInDoc(ctx, seriesStartUrl, r.HttpGetHook)
	if err != nil {
		return publication.Series{}, err
	}

	series := publication.Series{}
	info := doc.Find(".detail_header .info")
	series.Tags = []publication.Tag{
		{
			Value:   info.Find(".genre").Text(),
			IsGenre: true,
		},
	}
	series.Title = info.Find(".subj").Text()
	series.People = extractAuthors(info.Find(".author_area"))

	detail := doc.Find(".detail")
	series.Description = detail.Find(".summary").Text()

	if strings.Contains(detail.Find(".day_info").Text(), "COMPLETED") {
		series.Status = publication.StatusCompleted
	}

	series.Chapters = append(series.Chapters, extractChapters(doc)...)

	pages := utils.Filter(goquery.Map(doc.Find(".paginate a"), href), notEmpty)
	for index := 1; len(pages) > index; index++ {
		doc, err = r.httpClient.WrapInDoc(ctx, Domain+pages[index], r.HttpGetHook)
		if err != nil {
			return publication.Series{}, err
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

func (r *repository) HttpGetHook(req *http.Request) error {
	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")
	return nil
}

func searchUrl(keyword string) string {
	keyword = strings.TrimSpace(rg.ReplaceAllString(keyword, " "))
	return fmt.Sprintf(SearchUrl, url.QueryEscape(keyword))
}

func extractAuthors(sel *goquery.Selection) []publication.Person {
	sel.Find("button").Remove()

	authors := goquery.Map(sel.Find("a"), func(_ int, s *goquery.Selection) string {
		return s.Text()
	})

	plainTextAuthors := strings.ReplaceAll(sel.Text(), "...", "")
	authors = append(authors, strings.Split(plainTextAuthors, ",")...)

	return utils.Map(authors, func(s string) publication.Person {
		return publication.Person{
			Name: strings.TrimSpace(s),
		}
	})
}

func extractChapters(doc *goquery.Document) []publication.Chapter {
	return goquery.Map(doc.Find("#_listUl li a"), func(_ int, s *goquery.Selection) publication.Chapter {
		chapter := publication.Chapter{}
		chapter.Url = s.AttrOr("href", "")
		chapter.CoverUrl = s.Find("span img").AttrOr("src", "")
		chapter.Title = s.Find(".subj span").Text()
		chapter.Chapter = func() string {
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
