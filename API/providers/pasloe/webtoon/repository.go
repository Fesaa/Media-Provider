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
		log:        log.With().Str("handler", "webtoon-repository").Logger(),
	}
}

func (r *repository) Search(ctx context.Context, options SearchOptions) ([]SearchData, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, fmt.Sprintf(SearchUrl, url.QueryEscape(options.Query)) , r.HttpGetHook)
	if err != nil {
		return nil, err
	}

	var results []SearchData
	results = append(results, goquery.Map(doc.Find(".webtoon_list li"), r.extractSeries)...)
	return results, nil
}

func (r *repository) extractSeries(_ int, s *goquery.Selection) SearchData {
	rating := s.Find("a").First().AttrOr("data-title-unsuitable-for-children", "false")
	link := s.Find(".link._card_item").AttrOr("href", "")


	return SearchData{
		Id:              strings.TrimPrefix(link, Domain),
		Name:            s.Find(".title").Text(),
		ReadCount:       s.Find(".view_count").Text(),
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
	seriesStartUrl := Domain + id
	doc, err := r.httpClient.WrapInDoc(ctx, seriesStartUrl, r.HttpGetHook)
	if err != nil {
		return publication.Series{}, err
	}

	series := publication.Series{
		Id: id,
		RefUrl: seriesStartUrl,
	}
	info := doc.Find(".detail_header .info")
	series.Tags = []publication.Tag{
		{
			Value:   info.Find(".genre").Text(),
			IsGenre: true,
		},
	}
	series.Title = strings.Trim(info.Find(".subj").Text(), "\n\t")
	series.People = extractAuthors(info.Find(".author_area"))

	detail := doc.Find(".detail")
	series.Description = detail.Find(".summary").Text()

	if strings.Contains(detail.Find(".day_info").Text(), "COMPLETED") {
		series.Status = publication.StatusCompleted
	}

	series.Chapters = append(series.Chapters, extractChapters(doc)...)

	pages := utils.Filter(goquery.Map(doc.Find(".paginate a"), href), notEmpty)
	r.log.Trace().Int("pages", len(pages)).Msg("traversing pages")

	for index := 1; len(pages) > index; index++ {
		pageUrl := Domain+pages[index]
		r.log.Trace().Str("page", pageUrl).Msg("fetching page")

		doc, err = r.httpClient.WrapInDoc(ctx, pageUrl, r.HttpGetHook)
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
	return goquery.Map(doc.Find("._episodeItem > a"), func(_ int, s *goquery.Selection) publication.Chapter {
		chapterUrl := s.AttrOr("href", "")
		chapter := publication.Chapter{
			Id:          s.AttrOr("data-episode-no", chapterUrl),
			Title:       s.Find(".subj span").Text(),
			Chapter:     func() string {
				num := s.Find(".tx").Text()
				if len(num) > 0 && num[0] == '#' {
					return num[1:]
				}
				return num
			}(),
			CoverUrl:    s.Find("span img").AttrOr("src", ""),
			Url:         chapterUrl,
		}
		return chapter
	})
}

func notEmpty(s string) bool {
	return s != ""
}

func href(_ int, selection *goquery.Selection) string {
	return selection.AttrOr("href", "")
}
