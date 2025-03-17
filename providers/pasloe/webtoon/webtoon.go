package webtoon

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"io"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func NewWebToon(scope *dig.Scope) api.Downloadable {
	var wt *webtoon

	utils.Must(scope.Invoke(func(
		req payload.DownloadRequest, httpClient *http.Client,
		repository Repository, markdownService services.MarkdownService,
		fs afero.Afero,
	) {
		wt = &webtoon{
			id:              req.Id,
			httpClient:      httpClient,
			repository:      repository,
			markdownService: markdownService,
			fs:              fs,
		}

		wt.DownloadBase = api.NewDownloadableFromBlock[Chapter](scope, "webtoon", wt)
	}))
	return wt
}

type webtoon struct {
	httpClient      *http.Client
	repository      Repository
	markdownService services.MarkdownService
	fs              afero.Afero

	*api.DownloadBase[Chapter]
	id string

	searchInfo *SearchData
	info       *Series
}

func (w *webtoon) Title() string {
	if w.searchInfo != nil {
		return w.searchInfo.Name
	}
	if w.info != nil {
		return w.info.Name
	}

	if w.Req.TempTitle != "" {
		return w.Req.TempTitle
	}

	return w.id
}

func (w *webtoon) Provider() models.Provider {
	return w.Req.Provider
}

func (w *webtoon) RefUrl() string {
	if w.searchInfo != nil {
		return w.searchInfo.Url()
	}

	return ""
}

func (w *webtoon) LoadInfo(ctx context.Context) chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)
		info, err := w.repository.SeriesInfo(ctx, w.id)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				w.Log.Error().Err(err).Msg("error while loading webtoon info")
			}
			w.Cancel()
			return
		}

		w.info = info

		// TempTitle is the title we previously got from the search, just should ensure we get the correct stuff
		// WebToons search is surprisingly bad at correcting for spaces, special characters, etc...
		search, err := w.repository.Search(ctx, SearchOptions{Query: w.Req.TempTitle})
		if err != nil {
			w.Log.Error().Err(err).Msg("error while loading webtoon info")
			w.Cancel()
			return
		}

		w.searchInfo = utils.Find(search, func(data SearchData) bool {
			return data.Id == w.id
		})
		if w.searchInfo == nil {
			w.Log.Warn().Msg("was unable to load searchInfo, some meta-data may be off")
		}
	}()
	return out
}

func (w *webtoon) All() []Chapter {
	return w.info.Chapters
}

func (w *webtoon) ContentList() []payload.ListContentData {
	if w.info == nil {
		return nil
	}

	return utils.Map(w.info.Chapters, func(chapter Chapter) payload.ListContentData {
		return payload.ListContentData{
			SubContentId: chapter.Number,
			Selected:     len(w.ToDownloadUserSelected) == 0 || slices.Contains(w.ToDownloadUserSelected, chapter.Number),
			Label:        fmt.Sprintf("%s #%s - %s", w.info.Name, chapter.Number, chapter.Title),
		}
	})
}

func (w *webtoon) ContentDir(chapter Chapter) string {
	return fmt.Sprintf("%s Ch. %s", w.Title(), chapter.Number)
}

func (w *webtoon) ContentPath(chapter Chapter) string {
	return path.Join(w.Client.GetBaseDir(), w.GetBaseDir(), w.Title(), w.ContentDir(chapter))
}

func (w *webtoon) ContentKey(chapter Chapter) string {
	return chapter.Number
}

func (w *webtoon) ContentLogger(chapter Chapter) zerolog.Logger {
	return w.Log.With().Str("number", chapter.Number).Str("title", chapter.Title).Logger()
}

func (w *webtoon) ContentUrls(ctx context.Context, chapter Chapter) ([]string, error) {
	return w.repository.LoadImages(ctx, chapter)
}

func (w *webtoon) WriteContentMetaData(chapter Chapter) error {

	if w.Req.GetBool(IncludeCover, true) {
		// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
		// first by most readers, and in particular, kavita.
		filePath := path.Join(w.ContentPath(chapter), "!0000 cover.jpg")
		imageUrl := func() string {
			// Kavita uses the image of the first chapter as the cover image in lists
			// We replace this with the nicer looking image. As this software is still targeting Kavita
			if w.searchInfo != nil && chapter.Number == "1" {
				return webToonUrl(w.searchInfo.ThumbnailMobile)
			}
			return chapter.ImageUrl
		}()
		if err := w.downloadAndWrite(imageUrl, filePath); err != nil {
			return err
		}
	}

	w.Log.Trace().Str("chapter", chapter.Number).Msg("writing comicinfoxml")
	return comicinfo.Save(w.fs, w.comicInfo(), path.Join(w.ContentPath(chapter), "ComicInfo.xml"))
}

func (w *webtoon) comicInfo() *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = w.Title()
	ci.Summary = w.markdownService.SanitizeHtml(w.info.Description)
	ci.Manga = comicinfo.MangaYes
	ci.Genre = w.info.Genre

	if w.searchInfo != nil {
		ci.Writer = strings.Join(w.searchInfo.AuthorNameList, ",")
		ci.AgeRating = w.searchInfo.ComicInfoRating()
		ci.Web = w.searchInfo.Url()
	}

	if w.info.Completed {
		ci.Count = len(w.info.Chapters)
	}

	return ci
}

func (w *webtoon) DownloadContent(page int, chapter Chapter, url string) error {
	filePath := path.Join(w.ContentPath(chapter), fmt.Sprintf("page %s.jpg", utils.PadInt(page, 4)))
	if err := w.downloadAndWrite(url, filePath); err != nil {
		return err
	}
	w.ImagesDownloaded++
	return nil
}

var chapterRegex = regexp.MustCompile(".* Ch\\. (\\d+).cbz")

func (w *webtoon) IsContent(name string) bool {
	return chapterRegex.MatchString(name)
}

func (w *webtoon) ShouldDownload(chapter Chapter) bool {
	_, ok := w.GetContentByName(w.ContentDir(chapter) + ".cbz")
	return !ok
}

func (w *webtoon) downloadAndWrite(url string, path string, tryAgain ...bool) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			w.Log.Warn().Err(err).Msg("error closing body")
		}
	}(resp.Body)

	if resp.StatusCode == http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if err = w.fs.WriteFile(path, data, 0755); err != nil {
			return err
		}

		return nil
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	if len(tryAgain) > 0 && !tryAgain[0] {
		w.Log.Error().Msg("Reached rate limit, after sleeping. What is going on?")
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	retryAfter := resp.Header.Get("X-RateLimit-Retry-After")
	var d time.Duration
	if unix, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
		t := time.Unix(unix, 0)
		d = time.Until(t)
	} else {
		w.Log.Debug().Err(err).Str("retry-after", retryAfter).Msg("Could not parse retry-after")
		d = time.Minute
	}

	w.Log.Warn().Str("retryAfter", retryAfter).Dur("sleeping_for", d).Msg("Hit rate limit, try again after it's over")
	time.Sleep(d)
	return w.downloadAndWrite(url, path, false)
}

func webToonUrl(s string) string {
	return fmt.Sprintf("https://webtoon-phinf.pstatic.net%s", s)
}
