package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"log/slog"
	"os"
	"path"
	"sync"
)

var m MangadexClient

func I() MangadexClient {
	return m
}

func newClient(c MangadexConfig) MangadexClient {
	return &mangadexClientImpl{
		dir:         config.OrDefault(c.GetRootDir(), "temp"),
		maxImages:   c.GetMaxConcurrentMangadexImages(),
		mangas:      utils.NewSafeMap[string, Manga](),
		queue:       utils.NewQueue[payload.QueueStat](),
		downloading: nil,
		mu:          sync.Mutex{},
	}
}

type mangadexClientImpl struct {
	dir         string
	maxImages   int
	mangas      *utils.SafeMap[string, Manga]
	queue       utils.Queue[payload.QueueStat]
	downloading Manga
	mu          sync.Mutex
}

func (m *mangadexClientImpl) Download(req payload.DownloadRequest) (Manga, error) {
	if m.mangas.Has(req.Id) {
		return nil, fmt.Errorf("manga already exists: %s", req.Id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.downloading != nil {
		m.queue.Enqueue(req.ToQueueStat())
		return nil, nil
	}

	manga := newManga(req, m.maxImages, m)
	m.mangas.Set(req.Id, manga)
	m.downloading = manga
	manga.WaitForInfoAndDownload()
	return manga, nil
}

func (m *mangadexClientImpl) RemoveDownload(req payload.StopRequest) error {
	manga, ok := m.mangas.Get(req.Id)
	if !ok {
		ok = m.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == req.Id
		})
		if ok {
			slog.Info("manga removed from queue", "id", req.Id)
			return nil
		}
		return fmt.Errorf("manga not found: %s", req.Id)
	}

	slog.Info("Dropping manga", "title", manga.Title(), "id", manga.Id(), "deleteFiles", req.DeleteFiles)
	go func() {
		m.mangas.Delete(req.Id)
		manga.Cancel()
		m.mu.Lock()
		m.downloading = nil
		m.mu.Unlock()

		if req.DeleteFiles {
			go m.deleteFiles(manga)
		} else {
			go m.cleanup(manga)
		}
		m.startNext()
	}()
	return nil
}

func (m *mangadexClientImpl) startNext() {
	if m.queue.IsEmpty() {
		return
	}

	added := false
	for !added && !m.queue.IsEmpty() {
		q, _ := m.queue.Dequeue()
		_, err := m.Download(q.ToDownloadRequest())
		if err != nil {
			slog.Warn("Error while downloading manga from queue", "error", err)
			continue
		}
		added = true
	}
}

func (m *mangadexClientImpl) deleteFiles(manga Manga) {
	dir := path.Join(m.dir, manga.GetBaseDir(), manga.Title())
	slog.Info("Deleting directory", "dir", dir, "id", manga.Id())
	if err := os.RemoveAll(dir); err != nil {
		slog.Error("Error deleting directory", "dir", dir, "id", manga.Id(), "error", err)
	}
}

func (m *mangadexClientImpl) cleanup(manga Manga) {
	dir := path.Join(m.dir, manga.GetBaseDir(), manga.Title())
	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Error("Error reading directory", "dir", dir, "error", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		err := utils.ZipFolder(path.Join(dir, entry.Name()), path.Join(dir, entry.Name()+".cbz"))
		if err != nil {
			slog.Error("Error zipping directory", "dir", dir, "error", err, "id", manga.Id())
			continue
		}
		err = os.RemoveAll(path.Join(dir, entry.Name()))
		if err != nil {
			slog.Error("Error removing file", "file", entry.Name(), "error", err, "id", manga.Id())
			return
		}
	}
}

func (m *mangadexClientImpl) GetBaseDir() string {
	return m.dir
}

func (m *mangadexClientImpl) GetCurrentManga() Manga {
	return m.downloading
}

func (m *mangadexClientImpl) GetQueuedMangas() []payload.QueueStat {
	return m.queue.Items()
}
