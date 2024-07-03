package mangadex

import (
	"context"
	"errors"
	"fmt"
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
	"sync"
	"time"
)

var volumeRegex = regexp.MustCompile(".* Vol\\. (\\d+).cbz")

type mangaImpl struct {
	client MangadexClient
	log    *log.Logger

	id        string
	baseDir   string
	tempTitle string
	maxImages int

	info              *MangaSearchData
	chapters          ChapterSearchResponse
	covers            *utils.SafeMap[string, string]
	alreadyDownloaded []string

	chaptersDownloaded int
	imagesDownloaded   int
	lastTime           time.Time
	lastRead           int

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func newManga(req payload.DownloadRequest, maxImages int, client MangadexClient) Manga {
	manga := &mangaImpl{
		client:             client,
		id:                 req.Id,
		baseDir:            req.BaseDir,
		tempTitle:          req.TempTitle,
		maxImages:          min(maxImages, 4),
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

func (m *mangaImpl) Id() string {
	return m.id
}

func (m *mangaImpl) Title() string {
	if m.info == nil {
		return m.id
	}

	return m.info.Attributes.EnTitle()
}

func (m *mangaImpl) GetBaseDir() string {
	return m.baseDir
}

func (m *mangaImpl) GetDownloadDir() string {
	title := m.Title()
	if title == "" {
		return ""
	}
	return path.Join(m.baseDir, title)
}

func (m *mangaImpl) GetPrevVolumes() []string {
	return m.alreadyDownloaded
}

func (m *mangaImpl) GetInfo() payload.InfoStat {
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
		Progress:    utils.Percent(int64(m.chaptersDownloaded), int64(len(m.chapters.Data))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: m.GetDownloadDir(),
	}
}

func (m *mangaImpl) Cancel() {
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

func (m *mangaImpl) WaitForInfoAndDownload() {
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
			alreadyDownloaded, err := m.checkVolumesOnDisk()
			if err != nil {
				m.log.Warn("unable to check volumes on disk, downloading everything", "err", err)
			}
			m.alreadyDownloaded = alreadyDownloaded
			m.startDownload()
		}
	}()
}

func (m *mangaImpl) loadInfo() chan struct{} {
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
			m.log.Warn("error while loading manga covers, ignoring", "err", err)
			m.covers = &utils.SafeMap[string, string]{}
		} else {
			m.covers = utils.NewSafeMap(covers.GetUrlsPerVolume(m.id))
		}

		close(out)
	}()
	return out
}

func (m *mangaImpl) checkVolumesOnDisk() ([]string, error) {
	m.log.Debug("checking for already downloaded volumes", "mangaId", "dir", m.GetDownloadDir())
	entries, err := os.ReadDir(path.Join(m.client.GetBaseDir(), m.GetDownloadDir()))
	if errors.Is(err, os.ErrNotExist) {
		m.log.Debug("manga directory not found, fresh download")
		return []string{}, nil
	}
	if err != nil {
		return []string{}, err
	}

	out := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
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
	return out, nil
}

func (m *mangaImpl) startDownload() {
	m.log.Trace("starting download", "chapters", len(m.chapters.Data))
	m.wg = &sync.WaitGroup{}
	for _, chapter := range m.chapters.Data {
		if slices.Contains(m.alreadyDownloaded, m.volumeDir(chapter.Attributes.Volume)+".cbz") {
			m.log.Debug("skipping chapter, as the volume already exists",
				"volume", chapter.Attributes.Volume,
				"chapter", chapter.Attributes.Chapter)
			continue
		}

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

func (m *mangaImpl) downloadChapter(chapter ChapterSearchData) error {
	m.log.Trace("downloading chapter", "chapterId", chapter.Id, "chapterTitle", chapter.Attributes.Title)
	err := os.MkdirAll(m.chapterPath(chapter), 0755)
	if err != nil {
		return err
	}

	if err = m.tryVolumeCover(chapter); err != nil {
		m.log.Info("error while downloading cover image", "volume", chapter.Attributes.Volume, "err", err)
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
				go func() { <-sem }()
				if err = m.downloadImage(i, chapter, url); err != nil {
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

// TODO: don't try each chapter, only try once per volume
func (m *mangaImpl) tryVolumeCover(chapter ChapterSearchData) error {
	coverUrl, ok := m.covers.Get(chapter.Attributes.Volume)
	if !ok {
		m.log.Debug("unable to find cover", "volume", chapter.Attributes.Volume, "chapter", chapter.Attributes.Chapter)
		return nil
	}

	err := os.MkdirAll(m.volumePath(chapter), 0755)
	if err != nil {
		return err
	}

	m.log.Trace("downloading cover image", "volume", chapter.Attributes.Volume, "chapter", chapter.Attributes.Chapter, "url", coverUrl)
	resp, err := http.Get(coverUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	filePath := path.Join(m.volumePath(chapter), "cover.jpg")
	if err = os.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	return nil
}

func (m *mangaImpl) downloadImage(index int, chapter ChapterSearchData, url string) error {
	m.log.Trace("downloading image", "chapter", chapter.Attributes.Chapter, "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	filePath := path.Join(m.chapterPath(chapter), fmt.Sprintf("page %d.jpg", index))
	if err := os.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	m.imagesDownloaded++
	return nil
}

func (m *mangaImpl) mangaPath() string {
	return path.Join(m.client.GetBaseDir(), m.baseDir, m.Title())
}

func (m *mangaImpl) volumeDir(v string) string {
	return fmt.Sprintf("%s Vol. %s", m.Title(), v)
}

func (m *mangaImpl) volumePath(c ChapterSearchData) string {
	return path.Join(m.mangaPath(), m.volumeDir(c.Attributes.Volume))
}

func (m *mangaImpl) chapterPath(c ChapterSearchData) string {
	chDir := fmt.Sprintf("%s Vol. %s Ch. %s", m.Title(), c.Attributes.Volume, c.Attributes.Chapter)
	return path.Join(m.volumePath(c), chDir)
}
