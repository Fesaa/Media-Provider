package bato

import (
	"context"
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
)

const (
	Domain = "https://bato.to"

	QueryTag          = "word"
	GenresTag         = "genres"
	OriginalLangTag   = "orig"
	TranslatedLangTag = "lang"
	StatusTag         = "status"
	UploadTag         = "upload"
)

var (
	VolumeChapterRegex = regexp.MustCompile(`(?:Volume (\d+)\s+)?Chapter (\d+)`)
)

type Repository interface {
	Search(ctx context.Context, options SearchOptions) ([]SearchResult, error)
	SeriesInfo(ctx context.Context, id string) (*Series, error)
	ChapterImages(ctx context.Context, id string) ([]string, error)
}

func NewRepository(httpClient *http.Client, logger zerolog.Logger) Repository {
	return &repository{
		httpClient: httpClient,
		log:        logger,
	}
}

type repository struct {
	httpClient *http.Client
	log        zerolog.Logger
}

func (r *repository) Search(ctx context.Context, options SearchOptions) ([]SearchResult, error) {
	doc, err := r.wrapInDoc(ctx, searchUrl(options))
	if err != nil {
		return nil, err
	}

	series := doc.Find("div.grid.grid-cols-1.gap-5.border-t.border-t-base-200.pt-5 > div")
	return goquery.Map(series, r.selectionToSearchResult), nil
}

func (r *repository) selectionToSearchResult(_ int, sel *goquery.Selection) SearchResult {
	sr := SearchResult{}

	sr.Id = strings.TrimPrefix(sel.Find("div > a").First().AttrOr("href", ""), "/title/")
	sr.ImageUrl = sel.Find("div > a > img").First().AttrOr("src", "")

	info := sel.Find("div:nth-child(2)")
	sr.Title = info.Find("h3 a span span").First().Text()
	// sr.Authors = extractSeperatedList(info.Find("div:nth-child(2) span"), "/")
	// sr.Tags = extractSeperatedList(info.Find("div:nth-child(4) span"), ",")

	// meta := info.Find("div:nth-child(5)")
	// sr.LatestChapter = meta.Find("span a span").First().Text()
	// sr.UploaderImg = meta.Find("span div div a img").AttrOr("src", "")
	// sr.LastUploaded = meta.Find("span span time").First().Text()

	return sr
}

func extractSeperatedList(sel *goquery.Selection, sep string) []string {
	res := sel.Map(func(i int, s *goquery.Selection) string {
		return strings.TrimSpace(s.Text())
	})

	return utils.Filter(res, func(s string) bool {
		return s != sep
	})

}

func mapToContent(_ int, sel *goquery.Selection) string {
	return strings.TrimSpace(sel.Text())
}

func searchUrl(options SearchOptions) string {
	uri := utils.MustReturn(url.Parse(Domain + "/v3x-search"))
	q := uri.Query()
	q.Add(QueryTag, options.Query)

	if len(options.Genres) > 0 {
		q.Add(GenresTag, strings.Join(options.Genres, ","))
	}

	if len(options.OriginalLang) > 0 {
		q.Add(OriginalLangTag, strings.Join(options.OriginalLang, ","))
	}

	if len(options.TranslatedLang) > 0 {
		q.Add(TranslatedLangTag, strings.Join(options.TranslatedLang, ","))
	}

	if len(options.OriginalWorkStatus) > 0 {
		q.Add(StatusTag, strings.Join(utils.MapToString(options.OriginalWorkStatus), ","))
	}

	if len(options.BatoUploadStatus) > 0 {
		q.Add(UploadTag, strings.Join(utils.MapToString(options.BatoUploadStatus), ","))
	}

	uri.RawQuery = q.Encode()

	return uri.String()
}

func (r *repository) SeriesInfo(ctx context.Context, id string) (*Series, error) {
	doc, err := r.wrapInDoc(ctx, fmt.Sprintf("%s/title/%s", Domain, id))
	if err != nil {
		return nil, err
	}

	info := doc.Find("div.mt-3.grow.grid.gap-3.grid-cols-1")

	return &Series{
		Id:                id,
		Title:             info.Find("div > h3 a.link.link-hover").First().Text(),
		OriginalTitle:     info.Find("div > div > span").First().Text(),
		Authors:           info.Find("div.text-sm > a.link.link-hover.link-primary").Map(mapToContent),
		Tags:              extractSeperatedList(info.Find("div.space-y-2 > div.flex.items-center.flex-wrap > span > span"), ","),
		PublicationStatus: Publication(info.Find("div.space-y-2 > div > span.font-bold.uppercase").First().Text()),
		Summary:           info.Find("div.limit-html-p").First().Text(),
		WebLinks:          info.Find("div.limit-html div.limit-html-p a").Map(mapToContent),
		Chapters:          goquery.Map(doc.Find(`[name="chapter-list"] astro-slot > div`), r.readChapters),
	}, nil
}

func (r *repository) readChapters(i int, s *goquery.Selection) Chapter {
	chpt := Chapter{}

	uriEl := s.Find("div > a.link-hover.link-primary").First()
	chpt.Id = strings.TrimPrefix(uriEl.AttrOr("href", ""), "/title/")
	chpt.Volume, chpt.Chapter = extractVolumeAndChapter(uriEl.Text())
	chpt.Title = strings.TrimSpace(strings.TrimPrefix(s.Find("div > span").First().Text(), ": "))

	return chpt
}

func extractVolumeAndChapter(s string) (string, string) {
	matches := VolumeChapterRegex.FindStringSubmatch(s)

	if len(matches) == 0 {
		return "", ""
	}

	volume := ""
	if len(matches) > 1 {
		volume = matches[1]
	}
	chapter := ""
	if len(matches) > 2 {
		chapter = matches[2]
	}

	return volume, chapter
}

func (r *repository) ChapterImages(ctx context.Context, id string) ([]string, error) {
	doc, err := r.wrapInDoc(ctx, fmt.Sprintf("%s/title/%s", Domain, id))
	if err != nil {
		return nil, err
	}

	var imageProps ImageRenderProps
	islands := doc.Find(`astro-island[client="only"][component-export="default"]`)
	for i := range islands.Nodes {
		node := islands.Eq(i)
		props := node.AttrOr("props", "")
		if props == "" {
			continue
		}

		err = json.Unmarshal([]byte(props), &imageProps)
		if err != nil {
			continue
		}

		if len(imageProps.ImageFiles) == 0 {
			continue
		}
	}

	if len(imageProps.ImageFiles) != 2 {
		return nil, fmt.Errorf("no image props found for %s", id)
	}

	imagePropsString, ok := imageProps.ImageFiles[1].(string)
	if !ok {
		return nil, fmt.Errorf("no image props found for %s, was not a string", id)
	}

	var imageFiles [][]any
	if err = json.Unmarshal([]byte(imagePropsString), &imageFiles); err != nil {
		return nil, err
	}

	out := make([]string, len(imageFiles))
	for i := range imageFiles {
		if len(imageFiles[i]) < 2 {
			continue
		}

		out[i], ok = imageFiles[i][1].(string)
		if !ok {
			return nil, fmt.Errorf("no image props found for %s, was not a string", id)
		}
	}

	return out, nil
}

type ImageRenderProps struct {
	PageOpts   []any `json:"pageOpts"`
	ImageFiles []any `json:"imageFiles"`
	UrlP       []any `json:"urlP"`
}

func (r *repository) wrapInDoc(ctx context.Context, url string) (*goquery.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := r.httpClient.Do(req)
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
