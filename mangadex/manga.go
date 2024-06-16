package mangadex

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type mangaImpl struct {
	id                 string
	baseDir            string
	info               *MangaSearchData
	chapters           ChapterSearchResponse
	chaptersDownloaded int
	imagesDownloaded   int
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 *sync.WaitGroup

	lastTime time.Time
	lastRead int
}

func newManga(req payload.DownloadRequest) Manga {
	return &mangaImpl{
		id:                 req.Id,
		baseDir:            req.BaseDir,
		chaptersDownloaded: 0,
		imagesDownloaded:   0,
		lastRead:           0,
		lastTime:           time.Now(),
	}
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

func (m *mangaImpl) WaitForInfoAndDownload() {
	if m.cancel != nil {
		slog.Debug("manga already downloading", "id", m.id)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancel = cancel
	slog.Debug("Starting loading manga info", "id", m.id)
	go func() {
		select {
		case <-m.ctx.Done():
			return
		case <-m.loadInfo():
			slog.Info("Starting manga download", "id", m.id, "title", m.Title())
			m.startDownload()
		}
	}()
}

func (m *mangaImpl) loadInfo() chan struct{} {
	out := make(chan struct{})
	go func() {
		mangaInfo, err, _ := GetManga(m.id)
		if err != nil {
			slog.Error("An error occurred while loading manga info", "id", m.id, "err", err)
			m.cancel()
			return
		}
		m.info = &mangaInfo.Data

		chapters, err, _ := GetChapters(m.id)
		if err != nil || chapters == nil {
			slog.Error("An error occurred while getting chapters: ", err)
			m.cancel()
			return
		}
		m.chapters = chapters.FilterOneEnChapter()
		close(out)
	}()
	return out
}

func (m *mangaImpl) Cancel() {
	if m.cancel == nil {
		return
	}
	m.cancel()
	m.wg.Wait()
}

func (m *mangaImpl) startDownload() {
	m.wg = &sync.WaitGroup{}
	for _, chapter := range m.chapters.Data {
		select {
		case <-m.ctx.Done():
			m.wg.Wait()
			return
		default:
			m.wg.Add(1)
			err := m.downloadChapter(chapter)
			m.wg.Done()
			if err != nil {
				slog.Error("A fatal error occurred while downloading a chapter, cleaning up files", "id", m.id, "err", err)
				if err := I().RemoveDownload(payload.StopRequest{
					Provider:    config.MANGADEX,
					Id:          m.id,
					DeleteFiles: true,
				}); err != nil {
					slog.Error("Error cleaning up files", "id", m.id, "err", err)
				}
				m.wg.Wait()
				return
			}
		}

	}
	m.wg.Wait()
	if err := I().RemoveDownload(payload.StopRequest{
		Provider:    config.MANGADEX,
		Id:          m.id,
		DeleteFiles: false,
	}); err != nil {
		slog.Error("Error cleaning up files", "id", m.id, "err", err)
	}
}

func (m *mangaImpl) downloadChapter(chapter ChapterSearchData) error {
	slog.Debug("Downloading chapter", "id", m.id, "title", m.Title(), "chapter", m.chapterName(chapter))
	err := os.MkdirAll(path.Join(I().GetBaseDir(), m.baseDir, m.Title(), m.volumeName(chapter), m.chapterName(chapter)), 0755)
	if err != nil {
		return err
	}

	imageInfo, err, _ := GetChapterImages(chapter.Id)
	if err != nil {
		return err
	}
	urls := imageInfo.FullImageUrls()

	wg := sync.WaitGroup{}
	errCh := make(chan error, 1)
	sem := make(chan struct{}, 4)
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
				if err := m.downloadImage(i, chapter, url); err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
				}
			}(i, url)
		}

		if (i+1)%4 == 0 && i > 0 {
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

func (m *mangaImpl) downloadImage(index int, chapter ChapterSearchData, url string) error {
	//slog.Debug("Downloading image", "id", m.id, "chapter", m.chapterName(chapter), "url", url)
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

	filePath := path.Join(I().GetBaseDir(), m.baseDir, m.Title(), m.volumeName(chapter), m.chapterName(chapter), fmt.Sprintf("page %d.jpg", index))
	if err := os.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	m.imagesDownloaded++
	return nil
}

func (m *mangaImpl) volumeName(c ChapterSearchData) string {
	return fmt.Sprintf("%s Vol. %s", m.Title(), c.Attributes.Volume)
}

func (m *mangaImpl) chapterName(c ChapterSearchData) string {
	return fmt.Sprintf("%s Vol. %s Ch. %s", m.Title(), c.Attributes.Volume, c.Attributes.Chapter)
}

func (m *mangaImpl) Id() string {
	return m.id
}

func (m *mangaImpl) GetDownloadDir() string {
	title := m.Title()
	if title == "" {
		return ""
	}
	return path.Join(m.baseDir, title)
}

func (m *mangaImpl) GetInfo() payload.InfoStat {
	volumeDiff := m.imagesDownloaded - m.lastRead
	timeDiff := max(time.Since(m.lastTime).Seconds(), 1)
	speed := max(int64(float64(volumeDiff)/timeDiff), 1)
	m.lastRead = m.imagesDownloaded
	m.lastTime = time.Now()

	return payload.InfoStat{
		Provider:    config.MANGADEX,
		Id:          m.id,
		Name:        m.Title(),
		Size:        strconv.Itoa(len(m.chapters.Data)) + " Chapters",
		Progress:    utils.Percent(int64(m.chaptersDownloaded), int64(len(m.chapters.Data))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: m.GetDownloadDir(),
	}
}
