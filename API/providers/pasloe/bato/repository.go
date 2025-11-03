package bato

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog"
)

const (
	Domain = "https://bato.to"

	QueryTag          = "word"
	GenresTag         = "genres"
	IgnoredGenresTag  = "ignored_genres"
	OriginalLangTag   = "orig"
	TranslatedLangTag = "lang"
	StatusTag         = "status"
	UploadTag         = "upload"
)

// TODO: More flexibility with the default volume; Don't assign if no other volume has been found
type volumeChapterMapping struct {
	Regex         *regexp.Regexp
	DefaultVolume string
}

// TODO: Make these configurable? Especially mapping and cleans
//
//	Would be fun to have YAMLs for these. I think UI/DB is over the top for them
//	But maybe not?
var (
	VolumeChapterRegexes = []volumeChapterMapping{
		{regexp.MustCompile(`(?:(?:Volume|Vol\.?) ?(\d+)\s+)?(?:Chapter|Ch\.?) ([\d\\.]+)`), ""}, // Volume/Vol 1 Chapter/Ch 1.5
		{regexp.MustCompile(`(?:\[S(\d+)] ?)?Episode ([\d\\.]+)`), ""},                           // [S1] Episode 5
	}
	AuthorMappings = map[string]comicinfo.Roles{
		"(Story&Art)": {comicinfo.Writer, comicinfo.Colorist},
		"(Story)":     {comicinfo.Writer},
		"(Art)":       {comicinfo.Colorist},
	}
	TitleCleans = []string{
		"Official",
		"Unofficial",
		"Mature",
	}
	Braces = map[string]string{
		"[": "]",
		"(": ")",
		"«": "»",
	}
)

type Repository interface {
	Search(ctx context.Context, options SearchOptions) ([]SearchResult, error)
	SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error)
	ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error)
}

func NewRepository(httpClient *menou.Client, logger zerolog.Logger, markdown services.MarkdownService) Repository {
	return &repository{
		httpClient: httpClient,
		log:        logger.With().Str("handler", "bato-repository").Logger(),
		markdown:   markdown,
	}
}

type repository struct {
	httpClient *menou.Client
	log        zerolog.Logger
	markdown   services.MarkdownService
}

func (r *repository) Search(ctx context.Context, options SearchOptions) ([]SearchResult, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, searchUrl(options))
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
	if sr.Title == "" {
		sr.Title = info.Find("h3 a span").First().Text()
	}

	// sr.Authors = extractSeperatedList(info.Find("div:nth-child(2) span"), "/")
	// sr.Tags = extractSeperatedList(info.Find("div:nth-child(4) span"), ",")

	// meta := info.Find("div:nth-child(5)")
	// sr.LatestChapter = meta.Find("span a span").First().Text()
	// sr.UploaderImg = meta.Find("span div div a img").AttrOr("src", "")
	// sr.LastUploaded = meta.Find("span span time").First().Text()

	return sr
}

func extractSeperatedList(sel *goquery.Selection, sep string) []publication.Tag {
	res := goquery.Map(sel, func(i int, s *goquery.Selection) publication.Tag {
		return publication.Tag{
			Value: strings.TrimSpace(s.Text()),
		}
	})

	return utils.Filter(res, func(s publication.Tag) bool {
		return s.Value != sep
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

	if len(options.IgnoredGenres) > 0 {
		cur := q.Get(GenresTag)
		q.Set(GenresTag, cur+"|"+strings.Join(options.IgnoredGenres, ","))
	}

	if len(options.OriginalLang) > 0 {
		q.Add(OriginalLangTag, strings.Join(options.OriginalLang, ","))
	}

	if len(options.TranslatedLang) > 0 {
		q.Add(TranslatedLangTag, strings.Join(options.TranslatedLang, ","))
	}

	if len(options.OriginalWorkStatus) > 0 {
		q.Add(StatusTag, string(options.OriginalWorkStatus[0]))
	}

	if len(options.BatoUploadStatus) > 0 {
		q.Add(UploadTag, string(options.BatoUploadStatus[0]))
	}

	uri.RawQuery = q.Encode()

	return uri.String()
}

func (r *repository) SeriesInfo(ctx context.Context, id string, req payload.DownloadRequest) (publication.Series, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, fmt.Sprintf("%s/title/%s", Domain, id))
	if err != nil {
		return publication.Series{}, err
	}

	info := doc.Find("div.mt-3.grow.grid.gap-3.grid-cols-1")

	tss := utils.Settable[publication.Status]{}
	if ts := toPublicationStatus(strings.ToLower(info.Find("div.space-y-2 > div > span.font-bold.uppercase").Eq(1).Text())); ts != "" {
		tss.Set(ts)
	}

	return publication.Series{
		Id:                id,
		CoverUrl:          doc.Find("main > div > div > div > img").AttrOr("src", ""),
		Title:             cleanTitle(info.Find("div > h3 a.link.link-hover").First().Text()),
		AltTitle:          info.Find("div > div > span").First().Text(),
		People:            goquery.Map(info.Find("div.text-sm > a.link.link-hover.link-primary"), mapAuthor),
		Tags:              extractSeperatedList(info.Find("div.space-y-2 > div.flex.items-center.flex-wrap > span > span"), ","),
		Status:            toPublicationStatus(strings.ToLower(info.Find("div.space-y-2 > div > span.font-bold.uppercase").First().Text())),
		TranslationStatus: tss,
		Description:       r.markdown.SanitizeHtml(doc.Find(`meta[name="description"]`).First().AttrOr("content", "")),
		Links:             info.Find("div.limit-html div.limit-html-p a").Map(mapToContent),
		RefUrl:            fmt.Sprintf("%s/title/%s", Domain, id),
		Chapters:          goquery.Map(doc.Find(`[name="chapter-list"] astro-slot > div`), r.readChapters),
	}, nil
}

func cleanTitle(title string) string {
	for _, t := range TitleCleans {
		for start, end := range Braces {
			constructed := start + t + end
			title = strings.ReplaceAll(title, constructed, "")
		}
	}
	return strings.TrimSpace(title)
}

func mapAuthor(_ int, sel *goquery.Selection) publication.Person {
	cleaned := mapToContent(-1, sel)

	for v, roles := range AuthorMappings {
		if strings.Contains(cleaned, v) {
			return publication.Person{
				Name:  strings.ReplaceAll(cleaned, v, ""),
				Roles: roles,
			}
		}
	}

	return publication.Person{
		Name:  cleaned,
		Roles: comicinfo.Roles{comicinfo.Writer},
	}
}

func (r *repository) readChapters(_ int, s *goquery.Selection) publication.Chapter {
	chpt := publication.Chapter{}

	uriEl := s.Find("div > a.link-hover.link-primary").First()
	chpt.Id = strings.TrimPrefix(uriEl.AttrOr("href", ""), "/title/")
	chpt.Volume, chpt.Chapter = r.extractVolumeAndChapter(chpt.Id, uriEl.Text())

	chpt.Title = extractTitle(uriEl.Text())
	if chpt.Title == "" {
		titleText := s.Find("div > span.opacity-80").First().Text()
		chpt.Title = strings.TrimSpace(strings.TrimPrefix(titleText, ": "))
	}
	if chpt.Title == "" && chpt.Chapter == "" && chpt.Volume == "" {
		chpt.Title = strings.TrimSpace(uriEl.Text())
	}

	translatorEl := s.Find("div.avatar > div > a").First()
	if translatorEl != nil && translatorEl.AttrOr("href", "") != "" {
		chpt.Translator = []string{strings.TrimPrefix(translatorEl.AttrOr("href", ""), "/u/")}
	}

	return chpt
}

// extractVolumeAndChapter tries to get the volume(season) and chapter(episode). id is the chapter id for logging purposes
func (r *repository) extractVolumeAndChapter(id, s string) (string, string) {
	for _, mapping := range VolumeChapterRegexes {
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

func extractTitle(s string) string {
	idx := strings.Index(s, ":")
	if idx == -1 {
		return ""
	}

	if idx+1 == len(s) {
		return ""
	}

	return strings.TrimSpace(s[idx+1:])
}

func (r *repository) ChapterUrls(ctx context.Context, chapter publication.Chapter) ([]string, error) {
	doc, err := r.httpClient.WrapInDoc(ctx, fmt.Sprintf("%s/title/%s", Domain, chapter.Id))
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
			r.log.Trace().Err(err).Str("input", props).Msg("failed to unmarshal images")
			continue
		}

		if len(imageProps.ImageFiles) == 0 {
			r.log.Trace().Str("input", props).Msg("no images found in props, but was able to unmarshal")
			continue
		}
		break
	}

	if len(imageProps.ImageFiles) != 2 {
		return nil, fmt.Errorf("no image props found for %s", chapter.Id)
	}

	imagePropsString, ok := imageProps.ImageFiles[1].(string)
	if !ok {
		return nil, fmt.Errorf("no image props found for %s, was not a string", chapter.Id)
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
			return nil, fmt.Errorf("no image props found for %s, was not a string", chapter.Id)
		}
	}

	return out, nil
}

type ImageRenderProps struct {
	PageOpts   []any `json:"pageOpts"`
	ImageFiles []any `json:"imageFiles"`
	UrlP       []any `json:"urlP"`
}
