package dynasty

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
)

const (
	DOMAIN = "https://dynasty-scans.com"

	SEARCH  = DOMAIN + "/search?q=%s&classes[]=Series"
	SERIES  = DOMAIN + "/series/%s"
	CHAPTER = DOMAIN + "/chapters/%s"

	RELEASEDATAFORMAT = "Jan 2 '06"

	jsonOffset = 2
)

var (
	chapterTitleRegex        = regexp.MustCompile(`Chapter\s+([\d.]+)(?::\s*(.+))?`)
	chapterTitleRegexMatches = 3
)

type Repository interface {
	SearchSeries(ctx context.Context, options SearchOptions) ([]SearchData, error)
	SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error)
	ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error)
}

type repository struct {
	httpClient *menou.Client
	log        zerolog.Logger
}

func NewRepository(httpClient *menou.Client, log zerolog.Logger) Repository {
	return &repository{
		httpClient: httpClient,
		log:        log.With().Str("handler", "dynasty-repository").Logger(),
	}
}

func (r *repository) ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, chapterURL(chapter.Id))
	if err != nil {
		return nil, err
	}

	imageIds, err := r.extractImageIDs(doc.Find("script"))
	if err != nil {
		return nil, err
	}

	if len(imageIds) == 0 {
		return nil, fmt.Errorf("could not find chapter image")
	}

	urls := utils.Map(imageIds, func(id string) string {
		return DOMAIN + id
	})

	r.log.Trace().
		Str("chapterId", chapter.Id).
		Strs("images", urls).
		Int("amount", len(urls)).
		Msg("found chapter image ids")
	return urls, nil
}

func (r *repository) extractImageIDs(sel *goquery.Selection) ([]string, error) {
	var scriptContent string
	sel.Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "var pages") {
			scriptContent = text
		}
	})

	if scriptContent == "" {
		return nil, fmt.Errorf("could not find script")
	}

	start := strings.Index(scriptContent, "[{")
	end := strings.LastIndex(scriptContent, "}]") + jsonOffset
	if start == -1 || end == -1 {
		return nil, fmt.Errorf("could not find json data in script content")
	}
	jsonData := scriptContent[start:end]

	type Image struct {
		Path string `json:"image"`
	}

	var images []Image
	err := json.Unmarshal([]byte(jsonData), &images)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal json data: %w", err)
	}

	return utils.Map(images, func(t Image) string {
		return t.Path
	}), nil
}

func chapterURL(id string) string {
	return fmt.Sprintf(CHAPTER, id)
}

func (r *repository) SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, seriesURL(id))
	if err != nil {
		return publication.Series{}, err
	}

	series := publication.Series{
		Id:                id,
		Title:             doc.Find(".tag-title b").Text(),
		AltTitle:          doc.Find(".aliases b").Text(),
		Description:       doc.Find(".description p").Text(),
		CoverUrl:          DOMAIN + doc.Find(".thumbnail").AttrOr("src", ""),
		RefUrl:            "",
		Status:            toPublicationStatus(strings.TrimPrefix(doc.Find(".tag-title small").Last().Text(), "— ")),
		TranslationStatus: utils.Settable[publication.Status]{},
		Year:              0,
		OriginalLanguage:  "",
		ContentRating:     "",
		Tags:              utils.Filter(goquery.Map(doc.Find(".tag-tags a"), toTag), publication.NonEmptyTag),
		People:            nil,
		Links:             nil,
		Chapters:          r.readChapters(doc.Find(".chapter-list")),
	}

	return series, nil
}

func (r *repository) readChapters(chapterElement *goquery.Selection) []publication.Chapter {
	var chapters []publication.Chapter
	currentVolume := ""

	chapterElement.Children().Each(func(_ int, s *goquery.Selection) {
		if goquery.NodeName(s) == "dt" {
			if strings.Contains(s.Text(), "Volume") {
				currentVolume = strings.TrimPrefix(s.Text(), "Volume ")
			}
			return
		}

		if goquery.NodeName(s) != "dd" {
			r.log.Debug().Str("nodeName", goquery.NodeName(s)).Msg("skipping unknown html element in chapter list")
			return
		}

		titleElement := s.Find(".name")
		releaseDate := s.Find("small")
		tags := goquery.Map(s.Find(".label"), toTag)
		authors := goquery.Map(s.Find("a:not(.label):not(.name)"), toAuthor)

		chapter, title := func() (string, string) {
			chapterText := titleElement.Text()
			matches := chapterTitleRegex.FindStringSubmatch(chapterText)
			if len(matches) == chapterTitleRegexMatches {
				return matches[1], matches[2]
			}
			return "", chapterText
		}()

		releaseTime, err := time.Parse(RELEASEDATAFORMAT, strings.TrimPrefix(releaseDate.Text(), "released "))
		if err != nil {
			r.log.Warn().Err(err).Str("releaseDate", releaseDate.Text()).Msg("failed to parse release date")
		}

		chapters = append(chapters, publication.Chapter{
			Id:          strings.TrimPrefix(titleElement.AttrOr("href", ""), "/chapters/"),
			Title:       title,
			Volume:      currentVolume,
			Chapter:     chapter,
			ReleaseDate: &releaseTime,
			Tags:        tags,
			People:      authors,
		})
	})

	return chapters
}

func seriesURL(id string) string {
	return fmt.Sprintf(SERIES, id)
}

func (r *repository) SearchSeries(ctx context.Context, opt SearchOptions) ([]SearchData, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, searchURL(opt.Query))
	if err != nil {
		return nil, err
	}

	series := doc.Find(".chapter-list dd")
	return goquery.Map(series, r.selectionToSearchData), nil
}

func searchURL(keyword string) string {
	return fmt.Sprintf(SEARCH, url.QueryEscape(keyword))
}

func (r *repository) selectionToSearchData(_ int, sel *goquery.Selection) SearchData {
	sd := SearchData{}

	nameElement := sel.Find(".name").First()
	sd.Title = nameElement.Text()
	sd.Id = func() string {
		ref := nameElement.AttrOr("href", "")
		return strings.TrimPrefix(ref, "/series/")
	}()

	sd.Tags = utils.Filter(goquery.Map(sel.Find(".tags a"), toTag), publication.NonEmptyTag)
	sd.Authors = utils.Filter(goquery.Map(sel.Find("a"), toAuthor), func(person publication.Person) bool {
		return person.Name != ""
	})

	return sd
}

func toAuthor(_ int, s *goquery.Selection) publication.Person {
	const authorPrefix = "/authors/"
	if !strings.HasPrefix(s.AttrOr("href", ""), authorPrefix) {
		return publication.Person{}
	}

	href := strings.TrimPrefix(s.AttrOr("href", ""), authorPrefix)

	return publication.Person{
		Name: s.Text(),
		Url:  href,
	}
}

func toTag(_ int, s *goquery.Selection) publication.Tag {
	const tagPrefix = "/tags/"
	if !strings.HasPrefix(s.AttrOr("href", ""), tagPrefix) {
		return publication.Tag{}
	}

	href := strings.TrimPrefix(s.AttrOr("href", ""), tagPrefix)

	return publication.Tag{
		Value:      s.Text(),
		Identifier: href,
	}
}
