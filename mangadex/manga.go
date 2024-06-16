package mangadex

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type mangaImpl struct {
	client MangadexClient

	id        string
	baseDir   string
	tempTitle string
	maxImages int

	info     *MangaSearchData
	chapters ChapterSearchResponse
	covers   *utils.SafeMap[string, string]

	chaptersDownloaded int
	imagesDownloaded   int
	lastTime           time.Time
	lastRead           int

	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func newManga(req payload.DownloadRequest, maxImages int, client MangadexClient) Manga {
	return &mangaImpl{
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
}

func (m *mangaImpl) Title() string {
	if m.info == nil {
		return m.id
	}

	return m.info.Attributes.EnTitle()
}

func (m *mangaImpl) PrettyTitle() string {
	if m.info == nil {
		if m.tempTitle == "" {
			return m.id
		}
		return m.tempTitle
	}
	return m.info.Attributes.EnTitle()
}

func (m *mangaImpl) GetBaseDir() string {
	return m.baseDir
}

func (m *mangaImpl) WaitForInfoAndDownload() {
	if m.cancel != nil {
		log.Debug("manga already downloading", "mangaId", m.id, "title", m.Title())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.ctx = ctx
	m.cancel = cancel
	log.Trace("loading manga info", "mangaId", m.id)
	go func() {
		select {
		case <-m.ctx.Done():
			return
		case <-m.loadInfo():
			log.Debug("starting manga download", "mangaId", m.id, "title", m.Title())
			m.startDownload()
		}
	}()
}

func (m *mangaImpl) loadInfo() chan struct{} {
	out := make(chan struct{})
	go func() {
		mangaInfo, err := GetManga(m.id)
		if err != nil {
			log.Error("error while loading manga info", "mangaId", m.id, "err", err)
			m.cancel()
			return
		}
		m.info = &mangaInfo.Data

		chapters, err := GetChapters(m.id)
		if err != nil || chapters == nil {
			log.Error("error while loading manga chapters", "mangaId", m.id, "err", err)
			m.cancel()
			return
		}
		m.chapters = chapters.FilterOneEnChapter()

		covers, err := GetCoverImages(m.id)
		if err != nil || covers == nil {
			log.Warn("error while loading manga covers, ignoring", "mangaId", m.id, "err", err)
			m.covers = &utils.SafeMap[string, string]{}
		} else {
			m.covers = utils.NewSafeMap(covers.GetUrlsPerVolume(m.id))
		}

		close(out)
	}()
	return out
}

func (m *mangaImpl) Cancel() {
	log.Trace("calling cancel on manga", "mangaId", m.id)
	if m.cancel == nil {
		return
	}
	m.cancel()
	if m.wg == nil {
		return
	}
	m.wg.Wait()
}

func (m *mangaImpl) startDownload() {
	log.Trace("starting download", "mangaId", m.id, "title", m.Title(), "chapters", len(m.chapters.Data))
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
				log.Error("error while downloading a chapter, cleaning up", "mangaId", m.id, "title", m.Title(), "err", err)
				req := payload.StopRequest{
					Provider:    config.MANGADEX,
					Id:          m.id,
					DeleteFiles: true,
				}
				if err = m.client.RemoveDownload(req); err != nil {
					log.Error("error while cleaning up files", "mangaId", m.id, "title", m.Title(), "err", err)
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
		log.Error("error while cleaning up files", "mangaId", m.id, "title", m.Title(), "err", err)
	}
}

func (m *mangaImpl) downloadChapter(chapter ChapterSearchData) error {
	log.Trace("downloading chapter", "mangaId", m.id, "title", m.Title(), "chapterId", chapter.Id, "chapterTitle", chapter.Attributes.Title)
	err := os.MkdirAll(m.chapterDir(chapter), 0755)
	if err != nil {
		return err
	}

	if err = m.tryVolumeCover(chapter); err != nil {
		log.Info("error while downloading cover image", "mangaId", m.id, "volume", chapter.Attributes.Volume, "err", err)
	}

	imageInfo, err := GetChapterImages(chapter.Id)
	if err != nil {
		return err
	}
	urls := imageInfo.FullImageUrls()
	log.Trace("downloading images in chapter", "mangaId", m.id, "chapter", chapter.Attributes.Chapter, "images", len(urls))

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
				if err := m.downloadImage(i, chapter, url); err != nil {
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

func (m *mangaImpl) tryVolumeCover(chapter ChapterSearchData) error {
	coverUrl, ok := m.covers.Get(chapter.Attributes.Volume)
	if !ok {
		log.Debug("unable to find cover", "mangaId", m.id, "volume", chapter.Attributes.Volume, "chapter", chapter.Attributes.Chapter)
		return nil
	}

	err := os.MkdirAll(m.volumeDir(chapter), 0755)
	if err != nil {
		return err
	}

	log.Trace("downloading cover image", "mangaId", m.id, "volume", chapter.Attributes.Volume, "chapter", chapter.Attributes.Chapter, "url", coverUrl)
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

	filePath := path.Join(m.volumeDir(chapter), "cover.jpg")
	if err := os.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	return nil
}

func (m *mangaImpl) downloadImage(index int, chapter ChapterSearchData, url string) error {
	log.Trace("downloading image", "mangaId", m.id, "chapter", chapter.Attributes.Chapter, "url", url)
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

	filePath := path.Join(m.chapterDir(chapter), fmt.Sprintf("page %d.jpg", index))
	if err := os.WriteFile(filePath, data, 0755); err != nil {
		return err
	}

	m.imagesDownloaded++
	return nil
}

func (m *mangaImpl) mangaDir() string {
	return path.Join(m.client.GetBaseDir(), m.baseDir, m.Title())
}

func (m *mangaImpl) volumeDir(c ChapterSearchData) string {
	volDir := fmt.Sprintf("%s Vol. %s", m.Title(), c.Attributes.Volume)
	return path.Join(m.mangaDir(), volDir)
}

func (m *mangaImpl) chapterDir(c ChapterSearchData) string {
	chDir := fmt.Sprintf("%s Vol. %s Ch. %s", m.Title(), c.Attributes.Volume, c.Attributes.Chapter)
	return path.Join(m.volumeDir(c), chDir)
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
		Name:        m.PrettyTitle(),
		Size:        strconv.Itoa(len(m.chapters.Data)) + " Chapters",
		Downloading: m.wg != nil,
		Progress:    utils.Percent(int64(m.chaptersDownloaded), int64(len(m.chapters.Data))),
		SpeedType:   payload.IMAGES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: m.GetDownloadDir(),
	}
}
