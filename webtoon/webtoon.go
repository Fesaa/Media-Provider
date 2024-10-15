package webtoon

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/Fesaa/Media-Provider/wisewolf"
	"github.com/gofiber/fiber/v2"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

var chapterRegex = regexp.MustCompile(".* Ch\\. (\\d+).cbz")

func newWebToon(req payload.DownloadRequest, client Client) WebToon {
	wt := &webtoon{
		client:             client,
		id:                 req.Id,
		baseDir:            req.BaseDir,
		tempTitle:          req.TempTitle,
		maxImages:          min(c.GetConfig().GetMaxConcurrentMangadexImages(), 4),
		chaptersDownloaded: 0,
		imagesDownloaded:   0,
		lastTime:           time.Now(),
		lastRead:           0,
	}

	wt.log = log.With(slog.String("id", wt.id))
	return wt
}

type webtoon struct {
	client Client
	log    *log.Logger

	id        string
	baseDir   string
	tempTitle string
	maxImages int

	searchInfo    *SearchData
	info          *Series
	totalChapters int

	toDownload       []Chapter
	existingChapters []string

	chaptersDownloaded int
	imagesDownloaded   int
	lastTime           time.Time
	lastRead           int

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (w *webtoon) Title() string {
	if w.searchInfo != nil {
		return w.searchInfo.Name
	}
	if w.info != nil {
		return w.info.Name
	}

	if w.tempTitle != "" {
		return w.tempTitle
	}

	return w.id
}

func (w *webtoon) Id() string {
	return w.id
}

func (w *webtoon) GetBaseDir() string {
	return w.baseDir
}

func (w *webtoon) Downloading() bool {
	return w.wg != nil
}

func (w *webtoon) Cancel() {
	w.log.Trace("canceling webtoon download")
	if w.cancel == nil {
		return
	}
	w.cancel()
	if w.wg == nil {
		return
	}
	w.wg.Wait()
}

func (w *webtoon) WaitForInfoAndDownload() {
	if w.cancel != nil {
		w.log.Debug("webtoon has already started downloading")
		return
	}

	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.log.Trace("loading webtoon info")
	go func() {
		select {
		case <-w.ctx.Done():
			return
		case <-w.loadInfo():
			w.log = w.log.With(slog.String("title", w.Title()))
			w.log.Debug("starting webtoon download")
			w.checkChaptersOnDisk()
			w.startDownload()
		}
	}()
}

func (w *webtoon) loadInfo() chan struct{} {
	out := make(chan struct{})
	go func() {
		info, err := constructSeriesInfo(w.id)
		if err != nil {
			w.log.Error("error while loading webtoon info", "err", err)
			w.cancel()
			return
		}

		w.info = info

		search, err := Search(SearchOptions{Query: w.info.Name})
		if err != nil {
			w.log.Error("error while loading webtoon info", "err", err)
			w.cancel()
			return
		}

		w.searchInfo = utils.Find(search, func(data SearchData) bool {
			return fmt.Sprintf("%d", data.Id) == w.id
		})
		close(out)
	}()
	return out
}

func (w *webtoon) checkChaptersOnDisk() {
	w.log.Debug("checking already downloaded chapters", slog.String("dir", w.GetDownloadDir()))
	entries, err := os.ReadDir(path.Join(w.client.GetBaseDir(), w.GetDownloadDir()))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.log.Trace("directory not found, fresh download")
		} else {
			w.log.Warn("unable to check for downloaded chapters, downloading all", "error", err)
		}
		w.existingChapters = []string{}
		return
	}

	out := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cbz") {
			w.log.Trace("skipping non volume file", slog.String("file", entry.Name()))
			continue
		}

		matches := chapterRegex.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}
		w.log.Trace("found chapter on disk",
			slog.String("file", entry.Name()),
			slog.String("chapter", matches[1]))
		out = append(out, entry.Name())
	}

	w.log.Debug("found following chapters on disk", slog.String("chapters", fmt.Sprintf("%+v", out)))
	w.existingChapters = out
}

func (w *webtoon) startDownload() {
	w.log.Trace("starting download", slog.Int("chapters", len(w.info.Chapters)))
	w.wg = &sync.WaitGroup{}
	w.toDownload = utils.Filter(w.info.Chapters, func(chapter Chapter) bool {
		download := !slices.Contains(w.existingChapters, w.chapterDir(chapter.Number)+".cbz")
		if !download {
			w.log.Trace("chapter already downloaded, skipping", slog.String("chapter", chapter.Number))
		}
		return download
	})

	w.log.Debug("downloading chapters",
		slog.Int("all", len(w.info.Chapters)),
		slog.Int("toDownload", len(w.toDownload)))

	for _, chapter := range w.toDownload {
		select {
		case <-w.ctx.Done():
			w.wg.Wait()
			return
		default:
			w.wg.Add(1)
			err := w.downloadChapter(chapter)
			w.wg.Done()
			if err != nil {
				w.log.Error("error while downloading chapter, cleaning up", "err", err)
				req := payload.StopRequest{
					Provider:    config.WEBTOON,
					Id:          w.id,
					DeleteFiles: true,
				}
				if err = w.client.RemoveDownload(req); err != nil {
					w.log.Error("error while cleaning up", "err", err)
				}
				w.wg.Wait()
				return
			}
		}
	}

	w.wg.Wait()
	req := payload.StopRequest{
		Provider:    config.WEBTOON,
		Id:          w.id,
		DeleteFiles: false,
	}
	if err := w.client.RemoveDownload(req); err != nil {
		w.log.Error("error while cleaning up files", "err", err)
	}
}

func (w *webtoon) downloadChapter(chapter Chapter) error {
	l := w.log.With(slog.String("chapter", chapter.Number))

	l.Trace("downloading chapter")
	wg := &sync.WaitGroup{}

	err := os.MkdirAll(w.chapterPath(chapter.Number), 0755)
	if err != nil {
		return err
	}

	if err = w.writeChapterMetadata(chapter); err != nil {
		l.Warn("error while writing chapter metadata", "err", err)
	}

	urls, err := loadImages(chapter)
	if err != nil {
		return err
	}
	l.Debug("downloading images", "amount", len(urls))

	errCh := make(chan error, 1)
	sem := make(chan struct{}, w.maxImages)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, url := range urls {
		select {
		case <-w.ctx.Done():
			return nil
		case <-ctx.Done():
			wg.Wait()
			return errors.New("chapter download was cancelled from within")
		default:
			wg.Add(1)
			go func(i int, url string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				if err = w.downloadImage(i+1, chapter, url); err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
				}
			}(i, url)
		}

		if (i+1)%w.maxImages == 0 && i > 0 {
			select {
			case <-time.After(1 * time.Second):
			case err = <-errCh:
				wg.Wait()
				for len(sem) > 0 {
					<-sem
				}
				return fmt.Errorf("encountered an error while downloading images: %w", err)
			case <-ctx.Done():
				wg.Wait()
				return fmt.Errorf("chapter download was cancelled from within")
			}
		}

		select {
		case err = <-errCh:
			wg.Wait()
			for len(sem) > 0 {
				<-sem
			}
			return fmt.Errorf("encountered an error while downloading images: %w", err)
		default:
		}
	}

	wg.Wait()
	select {
	case err = <-errCh:
		return err
	default:
	}

	w.chaptersDownloaded++
	return nil
}

func (w *webtoon) downloadImage(page int, chapter Chapter, url string) error {
	filePath := path.Join(w.chapterPath(chapter.Number), fmt.Sprintf("page %s.jpg", padInt(page, 4)))
	if err := downloadAndWrite(url, filePath); err != nil {
		return err
	}
	w.imagesDownloaded++
	return nil
}

func (w *webtoon) writeChapterMetadata(chapter Chapter) error {
	// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
	// first by most readers, and in particular, kavita.
	filePath := path.Join(w.chapterPath(chapter.Number), "!0000 cover.jpg")
	imageUrl := func() string {
		// Kavita uses the image of the first chapter as the cover image in lists
		// We replace this with the nicer looking image. As this software is still targeting Kavita
		if w.searchInfo != nil && chapter.Number == "1" {
			return w.searchInfo.ThumbnailMobile
		}
		return chapter.ImageUrl
	}()
	if err := downloadAndWrite(imageUrl, filePath); err != nil {
		return err
	}

	w.log.Trace("writing comicinfoxml", "chapter", chapter.Number)
	return comicinfo.Save(w.comicInfo(), path.Join(w.chapterPath(chapter.Number), "ComicInfo.xml"))
}

func (w *webtoon) comicInfo() *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = w.Title()
	ci.Summary = w.info.Description
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

func (w *webtoon) GetInfo() payload.InfoStat {
	imageDiff := w.imagesDownloaded - w.lastRead
	timeDiff := max(time.Since(w.lastTime).Seconds(), 1)
	speed := max(int64(float64(imageDiff)/timeDiff), 1)
	w.lastRead = w.imagesDownloaded
	w.lastTime = time.Now()

	return payload.InfoStat{
		Provider: config.WEBTOON,
		Id:       w.id,
		Name:     w.Title(),
		Size: func() string {
			if w.info != nil {
				return strconv.Itoa(len(w.info.Chapters)) + " Chapters"
			}
			return ""
		}(),
		Downloading: w.wg != nil,
		Progress:    utils.Percent(int64(w.chaptersDownloaded), int64(len(w.toDownload))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: w.GetDownloadDir(),
	}
}

func (w *webtoon) GetDownloadDir() string {
	title := w.Title()
	if title == "" {
		return ""
	}
	return path.Join(w.baseDir, title)
}

func (w *webtoon) GetPrevChapters() []string {
	return w.existingChapters
}

func (w *webtoon) webToonPath() string {
	return path.Join(w.client.GetBaseDir(), w.baseDir, w.Title())
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

func padInt(i int, n int) string {
	return pad(strconv.Itoa(i), n)
}

func padFloat(f float64, n int) string {
	full := fmt.Sprintf("%.1f", f)
	parts := strings.Split(full, ".")
	if len(parts) < 2 || parts[1] == "0" { // No decimal part
		return pad(full, n)
	}
	return pad(parts[0], n) + "." + parts[1]
}

func pad(str string, n int) string {
	if len(str) < n {
		str = strings.Repeat("0", n-len(str)) + str
	}
	return str
}
