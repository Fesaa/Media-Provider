package dynasty

import (
	"encoding/json"
	"fmt"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
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
	chapterTitleRegex        = regexp.MustCompile(`Chapter\s+([\d.]+):\s*(.+)`)
	chapterTitleRegexMatches = 3
)

type Repository interface {
	SearchSeries(SearchOptions) ([]SearchData, error)
	SeriesInfo(id string) (*Series, error)
	ChapterImages(id string) ([]string, error)
}

type repository struct {
	httpClient *http.Client
	log        zerolog.Logger
}

func NewRepository(httpClient *http.Client, log zerolog.Logger) Repository {
	return &repository{
		httpClient: httpClient,
		log:        log.With().Str("handler", "dynasty-repository").Logger(),
	}
}

func (r *repository) ChapterImages(id string) ([]string, error) {
	doc, err := r.wrapInDoc(chapterURL(id))
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
		Str("chapterId", id).
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

func (r *repository) SeriesInfo(id string) (*Series, error) {
	doc, err := r.wrapInDoc(seriesURL(id))
	if err != nil {
		return nil, err
	}

	series := &Series{
		Id:          id,
		Title:       doc.Find(".tag-title b").Text(),
		AltTitle:    doc.Find(".aliases b").Text(),
		Description: doc.Find(".description p").Text(),
		CoverUrl:    DOMAIN + doc.Find(".thumbnail").AttrOr("src", ""),
		Status:      SeriesStatus(strings.TrimPrefix(doc.Find(".tag-title small").Text(), "— ")),
		Tags:        goquery.Map(doc.Find(".tag-tags a"), toTag),
		Authors:     goquery.Map(doc.Find(".tag-title a"), toAuthor),
		Chapters:    r.readChapters(doc.Find(".chapter-list")),
	}

	return series, nil
}

func (r *repository) readChapters(chapterElement *goquery.Selection) []Chapter {
	var chapters []Chapter
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

		chapters = append(chapters, Chapter{
			Id:          strings.TrimPrefix(titleElement.AttrOr("href", ""), "/chapters/"),
			Title:       title,
			Volume:      currentVolume,
			Chapter:     chapter,
			ReleaseDate: &releaseTime,
			Tags:        tags,
		})
	})

	return chapters
}

func seriesURL(id string) string {
	return fmt.Sprintf(SERIES, id)
}

func (r *repository) SearchSeries(opt SearchOptions) ([]SearchData, error) {
	doc, err := r.wrapInDoc(searchURL(opt.Query))
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
		if ref == "" {
			return ref
		}

		if strings.HasPrefix(ref, "/series/") {
			return strings.TrimPrefix(ref, "/series/")
		}
		return ref
	}()

	sd.Tags = goquery.Map(sel.Find(".tags a"), toTag)
	sd.Authors = goquery.Map(sel.Find("a"), toAuthor)

	return sd
}

func toAuthor(_ int, s *goquery.Selection) Author {
	href := strings.TrimPrefix(s.AttrOr("href", ""), "/authors/")

	return Author{
		DisplayName: s.Text(),
		Id:          href,
	}
}

func toTag(_ int, s *goquery.Selection) Tag {
	href := strings.TrimPrefix(s.AttrOr("href", ""), "/tags/")

	return Tag{
		DisplayName: s.Text(),
		Id:          href,
	}
}

func (r *repository) wrapInDoc(url string) (*goquery.Document, error) {
	res, err := r.httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			r.log.Warn().Err(err).Msg("failed to close body")
		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
