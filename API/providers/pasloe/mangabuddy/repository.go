package mangabuddy

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/providers/pasloe/bato"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const (
	timeFormat = "Jan, 2 2006"
	chapterUrl = "https://mangabuddy.com/api/manga/%s/chapters?source=detail"
)

var (
	bookIdRegex = regexp.MustCompile(`var\s+bookId\s*=\s*(\d+);`)
)

type Repository interface {
	Search(ctx context.Context, options SearchOptions) ([]payload.Info, error)
	SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error)
	ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]publication.DownloadUrl, error)
}

func NewRepository(httpClient *menou.Client, logger zerolog.Logger, markdown services.MarkdownService) Repository {
	return &repository{
		httpClient: httpClient,
		log:        logger.With().Str("handler", "mangabuffy-repository").Logger(),
		markdown:   markdown,
	}
}

type repository struct {
	httpClient *menou.Client
	log        zerolog.Logger
	markdown   services.MarkdownService
}

func (r *repository) Search(ctx context.Context, options SearchOptions) ([]payload.Info, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, searchUrl(options))
	if err != nil {
		return nil, err
	}

	series := doc.Find("div.book-detailed-item")
	return goquery.Map(series, func(_ int, s *goquery.Selection) payload.Info {
		return payload.Info{
			Name:        s.Find(".title > h3 > a").AttrOr("title", ""),
			Description: s.Find(".summary").Text(),
			Tags: goquery.Map(s.Find("div.genres > span"), func(_ int, s *goquery.Selection) payload.InfoTag {
				return payload.Of(s.Text(), s.AttrOr("class", ""))
			}),
			Size:     s.Find(".latest-chapter").Text(),
			Link:     domain + s.Find(".title > h3 > a").AttrOr("href", ""),
			InfoHash: s.Find(".title > h3 > a").AttrOr("href", ""),
			ImageUrl: s.Find(".thumb > a > img").AttrOr("data-src", ""), // TODO: Proxy
			RefUrl:   domain + s.Find(".title > h3 > a").AttrOr("href", ""),
			Provider: models.MANGA_BUDDY,
		}
	}), nil

}

func (r *repository) SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, domain+id)
	if err != nil {
		return publication.Series{}, err
	}

	meta := map[string][]string{}
	doc.Find("div.meta.box > p").Each(func(_ int, s *goquery.Selection) {
		key := strings.TrimSpace(strings.TrimSuffix(s.Find("strong").Text(), ":"))
		values := s.Find("a").Map(func(_ int, s *goquery.Selection) string {
			return strings.TrimSpace(strings.TrimSuffix(s.Text(), ","))
		})
		meta[key] = values
	})

	getFirstMeta := func(key string) string {
		values, ok := meta[key]
		if ok && len(values) > 0 {
			return values[0]
		}
		return ""
	}

	getMeta := func(key string) []string {
		values, ok := meta[key]
		if ok && len(values) > 0 {
			return values
		}
		return []string{}
	}

	script := doc.Find(".layout > script").First().Text()
	matches := bookIdRegex.FindStringSubmatch(script)
	if len(matches) < 1 {
		return publication.Series{}, errors.New("bookId not found in script")
	}

	chapters, err := r.httpClient.WrapInDoc(ctx, fmt.Sprintf(chapterUrl, matches[1]))
	if err != nil {
		return publication.Series{}, fmt.Errorf("failed to load chapters: %w", err)
	}

	return publication.Series{
		Id:    id,
		Title: doc.Find("div.detail > .name > h1").Text(),
		AltTitle: func() string {
			opt := strings.Split(doc.Find("div.detail > .name > h2").Text(), ";")
			if len(opt) > 0 {
				return opt[0]
			}
			return ""
		}(),
		Description:       strings.Trim(doc.Find("div.summary > p.content").Text(), "\n \r"),
		CoverUrl:          doc.Find(".img-cover > img").AttrOr("data-src", ""),
		RefUrl:            domain + id,
		Status:            publication.Status(strings.ToLower(getFirstMeta("Status"))),
		TranslationStatus: utils.Settable[publication.Status]{},
		Tags: utils.Map(getMeta("Genres"), func(t string) publication.Tag {
			t = strings.Trim(t, "\n ,")
			return publication.Tag{
				Value:      t,
				Identifier: t,
				IsGenre:    true,
			}
		}),
		People: utils.Map(getMeta("Authors"), func(t string) publication.Person {
			return publication.Person{
				Name:  t,
				Roles: []comicinfo.Role{comicinfo.Writer},
			}
		}),
		Chapters: goquery.Map(chapters.Find("#chapter-list > li"), func(_ int, s *goquery.Selection) publication.Chapter {
			chptId := s.Find("a").AttrOr("href", "")
			volume, chpt := r.extractVolumeAndChapter(chptId, s.Find("strong.chapter-title").Text())

			return publication.Chapter{
				Id:      chptId,
				Title:   s.Find("a").AttrOr("title", ""),
				Volume:  volume,
				Chapter: chpt,
				Url:     domain + id,
			}
		}),
	}, nil
}

// extractVolumeAndChapter tries to get the volume(season) and chapter(episode). id is the chapter id for logging purposes
func (r *repository) extractVolumeAndChapter(id, s string) (string, string) {
	for _, mapping := range bato.VolumeChapterRegexes {
		matches := mapping.Regex.FindStringSubmatch(s)

		if len(matches) == 0 {
			continue
		}

		volume := ""
		if len(matches) > 1 {
			volume = matches[1]
		}
		chapter := ""
		if len(matches) > 2 {
			chapter = matches[2]
		}

		return utils.OrElse(volume, mapping.DefaultVolume), chapter
	}

	r.log.Trace().Str("chapter", id).Str("input", s).Msg("failed to match volume and chapter")
	return "", ""
}

func (r *repository) ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]publication.DownloadUrl, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, domain+chapter.Id)
	if err != nil {
		return nil, err
	}

	var scriptContent string
	doc.Find("script").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "var chapImages") {
			scriptContent = text
		}
	})

	scriptContent = strings.Trim(strings.ReplaceAll(scriptContent, "var chapImages = '", ""), "\n ';")
	return utils.Map(strings.Split(scriptContent, ","), publication.AsDownloadUrl), nil
}

func (r *repository) HttpGetHook(req *http.Request) error {
	req.Header.Add(fiber.HeaderReferer, "https://mangabuddy.com/")
	return nil
}

func searchUrl(options SearchOptions) string {
	searchUri := utils.MustReturn(url.Parse(domain + "/search"))
	q := searchUri.Query()
	q.Add("q", options.Query)

	for _, genre := range options.Genres {
		q.Add("genre[]", genre)
	}

	if len(options.Status) > 0 {
		q.Add("status", options.Status)
	}

	if len(options.OrderBy) > 0 {
		q.Add("sort", options.OrderBy)
	}

	searchUri.RawQuery = q.Encode()
	return searchUri.String()
}
