package mangadex

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/comicinfo"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
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

const comicInfoNote = "This comicinfo.xml was auto generated by Media-Provider, with information from mangadex. Source code can be found here: https://github.com/Fesaa/Media-Provider/"

var volumeRegex = regexp.MustCompile(".* Vol\\. (\\d+).cbz")

type manga struct {
	client Client
	log    *log.Logger

	id        string
	baseDir   string
	tempTitle string
	maxImages int

	info     *MangaSearchData
	chapters ChapterSearchResponse

	toDownload   []ChapterSearchData
	coverFactory CoverFactory

	volumeMetadata  []string
	existingVolumes []string

	chaptersDownloaded int
	imagesDownloaded   int
	lastTime           time.Time
	lastRead           int

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func newManga(req payload.DownloadRequest, c Config, client Client) Manga {
	manga := &manga{
		client:             client,
		id:                 req.Id,
		baseDir:            req.BaseDir,
		tempTitle:          req.TempTitle,
		maxImages:          min(c.GetMaxConcurrentMangadexImages(), 4),
		volumeMetadata:     make([]string, 0),
		chaptersDownloaded: 0,
		imagesDownloaded:   0,
		lastRead:           0,
		lastTime:           time.Now(),
		wg:                 nil,
	}

	manga.log = log.With(
		slog.String("mangaId", manga.id),
		slog.String("title", manga.Title()))
	return manga
}

func (m *manga) Id() string {
	return m.id
}

func (m *manga) Title() string {
	if m.info == nil {
		return m.id
	}

	return m.info.Attributes.EnTitle()
}

func (m *manga) GetBaseDir() string {
	return m.baseDir
}

func (m *manga) GetDownloadDir() string {
	title := m.Title()
	if title == "" {
		return ""
	}
	return path.Join(m.baseDir, title)
}

func (m *manga) GetPrevVolumes() []string {
	return m.existingVolumes
}

func (m *manga) GetInfo() payload.InfoStat {
	volumeDiff := m.imagesDownloaded - m.lastRead
	timeDiff := max(time.Since(m.lastTime).Seconds(), 1)
	speed := max(int64(float64(volumeDiff)/timeDiff), 1)
	m.lastRead = m.imagesDownloaded
	m.lastTime = time.Now()

	return payload.InfoStat{
		Provider: config.MANGADEX,
		Id:       m.id,
		Name: func() string {
			title := m.Title()
			if title == m.id && m.tempTitle != "" {
				return m.tempTitle
			}
			return title
		}(),
		Size:        strconv.Itoa(len(m.chapters.Data)) + " Chapters",
		Downloading: m.wg != nil,
		Progress:    utils.Percent(int64(m.chaptersDownloaded), int64(len(m.toDownload))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: m.GetDownloadDir(),
	}
}

func (m *manga) Cancel() {
	m.log.Trace("calling cancel on manga")
	if m.cancel == nil {
		return
	}
	m.cancel()
	if m.wg == nil {
		return
	}
	m.wg.Wait()
}

func (m *manga) WaitForInfoAndDownload() {
	if m.cancel != nil {
		m.log.Debug("manga already downloading")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancel = cancel
	m.log.Trace("loading manga info")
	go func() {
		select {
		case <-m.ctx.Done():
			return
		case <-m.loadInfo():
			m.log.Debug("starting manga download")
			m.checkVolumesOnDisk()
			m.startDownload()
		}
	}()
}

func (m *manga) loadInfo() chan struct{} {
	out := make(chan struct{})
	go func() {
		mangaInfo, err := GetManga(m.id)
		if err != nil {
			m.log.Error("error while loading manga info", "err", err)
			m.cancel()
			return
		}
		m.info = &mangaInfo.Data

		chapters, err := GetChapters(m.id)
		if err != nil || chapters == nil {
			m.log.Error("error while loading manga chapters", "err", err)
			m.cancel()
			return
		}
		m.chapters = chapters.FilterOneEnChapter()

		covers, err := GetCoverImages(m.id)
		if err != nil || covers == nil {
			m.log.Warn("error while loading manga coverFactory, ignoring", "err", err)
			m.coverFactory = func(volume string) (string, bool) {
				return "", false
			}
		} else {
			m.coverFactory = covers.GetCoverFactory(m.id)
		}

		close(out)
	}()
	return out
}

func (m *manga) checkVolumesOnDisk() {
	m.log.Debug("checking for already downloaded volumes", "mangaId", "dir", m.GetDownloadDir())
	entries, err := os.ReadDir(path.Join(m.client.GetBaseDir(), m.GetDownloadDir()))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.log.Debug("directory not found, fresh download")
		} else {
			m.log.Warn("unable to check for downloaded volumes, downloading all", "err", err)
		}
		m.existingVolumes = []string{}
		return
	}

	out := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cbz") {
			m.log.Trace("skipping non volume file", "file", entry.Name())
			continue
		}

		matches := volumeRegex.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}
		m.log.Trace("found volume on disk", "file", entry.Name(), "volume", matches[1])
		out = append(out, entry.Name())
	}
	m.log.Debug("found following volumes on disk", "volumes", fmt.Sprintf("%+v", out))
	m.existingVolumes = out
}

func (m *manga) startDownload() {
	m.log.Trace("starting download", "chapters", len(m.chapters.Data))
	m.wg = &sync.WaitGroup{}
	m.toDownload = utils.Filter(m.chapters.Data, func(chapter ChapterSearchData) bool {
		download := !slices.Contains(m.existingVolumes, m.volumeDir(chapter.Attributes.Volume)+".cbz")
		if !download {
			m.log.Trace("chapter already downloaded, skipping", "volume", chapter.Attributes.Volume, "chapter", chapter.Attributes.Chapter)
		}
		return download
	})

	m.log.Debug("downloading chapters", "all", len(m.chapters.Data), "toDownload", len(m.toDownload))
	for _, chapter := range m.toDownload {
		select {
		case <-m.ctx.Done():
			m.wg.Wait()
			return
		default:
			m.wg.Add(1)
			err := m.downloadChapter(chapter)
			m.wg.Done()
			if err != nil {
				m.log.Error("error while downloading a chapter, cleaning up", "err", err)
				req := payload.StopRequest{
					Provider:    config.MANGADEX,
					Id:          m.id,
					DeleteFiles: true,
				}
				if err = m.client.RemoveDownload(req); err != nil {
					m.log.Error("error while cleaning up files", "err", err)
				}
				m.wg.Wait()
				return
			}
		}

	}
	m.wg.Wait()
	req := payload.StopRequest{
		Provider:    config.MANGADEX,
		Id:          m.id,
		DeleteFiles: false,
	}
	if err := m.client.RemoveDownload(req); err != nil {
		m.log.Error("error while cleaning up files", "err", err)
	}
}

func (m *manga) downloadChapter(chapter ChapterSearchData) error {
	m.log.Trace("downloading chapter", "chapterId", chapter.Id, "chapterTitle", chapter.Attributes.Title)
	err := os.MkdirAll(m.chapterPath(chapter), 0755)
	if err != nil {
		return err
	}

	if err = m.writeVolumeMetadata(chapter); err != nil {
		m.log.Info("error while writing volume metadata", "volume", chapter.Attributes.Volume, "err", err)
	}

	imageInfo, err := GetChapterImages(chapter.Id)
	if err != nil {
		return err
	}
	urls := imageInfo.FullImageUrls()
	m.log.Trace("downloading images in chapter", "chapter", chapter.Attributes.Chapter, "images", len(urls))

	wg := sync.WaitGroup{}
	errCh := make(chan error, 1)
	sem := make(chan struct{}, m.maxImages)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, url := range urls {
		select {
		case <-m.ctx.Done():
			return nil
		case <-ctx.Done():
			wg.Wait()
			return fmt.Errorf("chapter download was cancelled from within")
		default:
			wg.Add(1)
			go func(i int, url string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				// Indexing pages from 1
				if err = m.downloadImage(i+1, chapter, url); err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
				}
			}(i, url)
		}

		if (i+1)%m.maxImages == 0 && i > 0 {
			select {
			case <-time.After(1 * time.Second):
			case err := <-errCh:
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
		case err := <-errCh:
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
	case err := <-errCh:
		return err
	default:
	}

	m.chaptersDownloaded++
	return nil
}

func (m *manga) writeVolumeMetadata(chapter ChapterSearchData) error {
	if slices.Contains(m.volumeMetadata, chapter.Attributes.Volume) {
		m.log.Debug("volume metadata already written, skipping", "volume", chapter.Attributes.Volume, "chapter", chapter.Attributes.Chapter)
		return nil
	}

	err := os.MkdirAll(m.volumePath(chapter), 0755)
	if err != nil {
		return err
	}

	coverUrl, ok := m.coverFactory(chapter.Attributes.Volume)
	if !ok {
		m.log.Debug("unable to find cover", "volume", chapter.Attributes.Volume)
	} else {
		m.log.Trace("downloading cover image", "volume", chapter.Attributes.Volume, "url", coverUrl)
		// Use !0000 cover.jpg to make sure it's the first file in the archive, this causes it to be read
		// first by most readers, and in particular, kavita.
		filePath := path.Join(m.volumePath(chapter), "!0000 cover.jpg")
		if err = downloadAndWrite(coverUrl, filePath); err != nil {
			return err
		}
	}

	m.log.Debug("writing comicinfoxml", "volume", chapter.Attributes.Volume)
	if err = comicinfo.Save(m.comicInfo(chapter), path.Join(m.volumePath(chapter), "comicinfo.xml")); err != nil {
		return err
	}

	m.volumeMetadata = append(m.volumeMetadata, chapter.Attributes.Volume)
	return nil
}

func (m *manga) comicInfo(chapter ChapterSearchData) *comicinfo.ComicInfo {
	ci := comicinfo.NewComicInfo()

	ci.Series = m.info.Attributes.EnTitle()
	ci.Year = m.info.Attributes.Year
	ci.Summary = m.info.Attributes.EnDescription()
	ci.Manga = comicinfo.MangaYes
	ci.AgeRating = m.info.Attributes.ContentRating.ComicInfoAgeRating()

	alts := m.info.Attributes.EnAltTitles()
	if len(alts) > 0 {
		ci.LocalizedSeries = alts[0]
	}

	if v, err := strconv.Atoi(chapter.Attributes.Volume); err == nil {
		ci.Volume = v
	} else {
		m.log.Trace("unable to parse volume number", "volume", chapter.Attributes.Volume, "err", err)
	}

	ci.Tags = strings.Join(utils.MaybeMap(m.info.Attributes.Tags, func(t TagData) (string, bool) {
		n, ok := t.Attributes.Name["en"]
		if !ok {
			return "", false
		}
		return n, true
	}), ",")

	ci.Notes = comicInfoNote
	return ci
}

func (m *manga) downloadImage(page int, chapter ChapterSearchData, url string) error {
	m.log.Trace("downloading image", "chapter", chapter.Attributes.Chapter, "url", url)
	filePath := path.Join(m.chapterPath(chapter), fmt.Sprintf("page %s.jpg", padNumber(page, 4)))
	if err := downloadAndWrite(url, filePath); err != nil {
		return err
	}
	m.imagesDownloaded++
	return nil
}

func (m *manga) mangaPath() string {
	return path.Join(m.client.GetBaseDir(), m.baseDir, m.Title())
}

func (m *manga) volumeDir(v string) string {
	return fmt.Sprintf("%s Vol. %s", m.Title(), v)
}

func (m *manga) volumePath(c ChapterSearchData) string {
	return path.Join(m.mangaPath(), m.volumeDir(c.Attributes.Volume))
}

func (m *manga) chapterPath(c ChapterSearchData) string {
	chDir := fmt.Sprintf("%s Vol. %s Ch. %s", m.Title(), c.Attributes.Volume, c.Attributes.Chapter)
	return path.Join(m.volumePath(c), chDir)
}

func downloadAndWrite(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
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

func padNumber(num int, n int) string {
	str := strconv.Itoa(num)
	if len(str) < n {
		str = strings.Repeat("0", n-len(str)) + str
	}
	return str
}
