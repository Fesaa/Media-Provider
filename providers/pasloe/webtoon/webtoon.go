package webtoon

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/http/wisewolf"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var chapterRegex = regexp.MustCompile(".* Ch\\. (\\d+).cbz")

func NewWebToon(req payload.DownloadRequest, client api.Client) api.Downloadable {
	wt := &webtoon{
		id: req.Id,
	}

	d := api.NewDownloadableFromBlock[Chapter](req, wt, client)
	wt.DownloadBase = d
	return wt
}

type webtoon struct {
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

func (w *webtoon) LoadInfo() chan struct{} {
	out := make(chan struct{})
	go func() {
		info, err := constructSeriesInfo(w.id)
		if err != nil {
			w.Log.Error("error while loading webtoon info", "err", err)
			w.Cancel()
			return
		}

		w.info = info

		// TempTitle is the title we previously got from the search, just should ensure we get the correct stuff
		// WebToons search is surprisingly bad at correcting for spaces, special characters, etc...
		search, err := Search(SearchOptions{Query: w.Req.TempTitle})
		if err != nil {
			w.Log.Error("error while loading webtoon info", "err", err)
			w.Cancel()
			return
		}

		w.searchInfo = utils.Find(search, func(data SearchData) bool {
			return fmt.Sprintf("%d", data.Id) == w.id
		})
		if w.searchInfo == nil {
			w.Log.Warn("was unable to load searchInfo, some meta-data may be off")
		}

		close(out)
	}()
	return out
}

func (w *webtoon) All() []Chapter {
	return w.info.Chapters
}

func (w *webtoon) GetInfo() payload.InfoStat {
	imageDiff := w.ImagesDownloaded - w.LastRead
	timeDiff := max(time.Since(w.LastTime).Seconds(), 1)
	speed := max(int64(float64(imageDiff)/timeDiff), 1)
	w.LastRead = w.ImagesDownloaded
	w.LastTime = time.Now()

	return payload.InfoStat{
		Provider: models.WEBTOON,
		Id:       w.id,
		Name:     w.Title(),
		Size: func() string {
			if w.info != nil {
				return strconv.Itoa(len(w.ToDownload)) + " Chapters"
			}
			return ""
		}(),
		Downloading: w.Wg != nil,
		Progress:    utils.Percent(int64(w.ContentDownloaded), int64(len(w.ToDownload))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: w.GetDownloadDir(),
	}
}

func (w *webtoon) ContentDir(chapter Chapter) string {
	return w.chapterDir(chapter.Number)
}

func (w *webtoon) ContentPath(chapter Chapter) string {
	return w.chapterPath(chapter.Number)
}

func (w *webtoon) ContentKey(chapter Chapter) string {
	return chapter.Number
}

func (w *webtoon) ContentLogger(chapter Chapter) *log.Logger {
	return w.Log.With("number", chapter.Number, "title", chapter.Title)
}

func (w *webtoon) ContentUrls(chapter Chapter) ([]string, error) {
	return loadImages(chapter)
}

func (w *webtoon) WriteContentMetaData(chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(w.chapterPath(chapter.Number), "!0000 cover.jpg")
	imageUrl := func() string {
		// Kavita uses the image of the first chapter as the cover image in lists
		// We replace this with the nicer looking image. As this software is still targeting Kavita
		if w.searchInfo != nil && chapter.Number == "1" {
			return webToonUrl(w.searchInfo.ThumbnailMobile)
		}
		return chapter.ImageUrl
	}()
	if err := downloadAndWrite(imageUrl, filePath); err != nil {
		return err
	}

	w.Log.Trace("writing comicinfoxml", "chapter", chapter.Number)
	return comicinfo.Save(w.comicInfo(), path.Join(w.chapterPath(chapter.Number), "ComicInfo.xml"))
}

func (w *webtoon) comicInfo() *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = w.Title()
	ci.Summary = utils.SanitizeHtml(w.info.Description)
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
	filePath := path.Join(w.chapterPath(chapter.Number), fmt.Sprintf("page %s.jpg", utils.PadInt(page, 4)))
	if err := downloadAndWrite(url, filePath); err != nil {
		return err
	}
	w.ImagesDownloaded++
	return nil
}

func (w *webtoon) ContentRegex() *regexp.Regexp {
	return chapterRegex
}

func (w *webtoon) webToonPath() string {
	return path.Join(w.Client.GetBaseDir(), w.GetBaseDir(), w.Title())
}

func (w *webtoon) chapterDir(number string) string {
	return fmt.Sprintf("%s Ch. %s", w.Title(), number)
}

func (w *webtoon) chapterPath(number string) string {
	return path.Join(w.webToonPath(), w.chapterDir(number))
}

func downloadAndWrite(url string, path string, tryAgain ...bool) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add(fiber.HeaderReferer, "https://www.webtoons.com/")

	resp, err := wisewolf.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusTooManyRequests {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		retryAfter := resp.Header.Get("X-RateLimit-Retry-After")
		if retryAfter == "" {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		if unix, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
			t := time.Unix(unix, 0)

			if len(tryAgain) > 0 && !tryAgain[0] {
				log.Error("Reached rate limit, after sleeping. What is going on?")
				return fmt.Errorf("bad status: %s", resp.Status)
			}

			d := time.Until(t)
			log.Warn("Hit rate limit, try again after it's over",
				slog.String("retryAfter", retryAfter),
				slog.Duration("sleeping_for", d))

			time.Sleep(d)
			return downloadAndWrite(url, path, false)
		}

	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			log.Warn("error while closing response body", "err", err)
		}
	}(resp.Body)
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, data, 0755); err != nil {
		return err
	}

	return nil
}

func webToonUrl(s string) string {
	return fmt.Sprintf("https://webtoon-phinf.pstatic.net%s", s)
}
