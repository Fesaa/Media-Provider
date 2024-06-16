package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
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

func newClient() MangadexClient {
	return &mangadexClientImpl{
		dir:         config.OrDefault(config.I().GetRootDir(), "temp"),
		mangas:      utils.NewSafeMap[string, Manga](),
		queue:       utils.NewQueue[config.QueueStat](),
		downloading: nil,
		mu:          sync.Mutex{},
	}
}

type mangadexClientImpl struct {
	dir         string
	mangas      *utils.SafeMap[string, Manga]
	queue       utils.Queue[config.QueueStat]
	downloading Manga
	mu          sync.Mutex
}

func (m *mangadexClientImpl) Download(id string, baseDir string) error {
	if m.mangas.Has(id) {
		return fmt.Errorf("manga already exists: %s", id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.downloading != nil {
		m.queue.Enqueue(config.QueueStat{
			Provider: config.MANGADEX,
			Id:       id,
			BaseDir:  baseDir,
		})
	} else {
		manga := newManga(id, baseDir)
		m.mangas.Set(id, manga)
		m.downloading = manga
		manga.WaitForInfoAndDownload()
	}

	return nil
}

func (m *mangadexClientImpl) RemoveDownload(id string, deleteFiles bool) error {
	manga, ok := m.mangas.Get(id)
	if !ok {
		ok = m.queue.RemoveFunc(func(item config.QueueStat) bool {
			return item.Id == id
		})
		if ok {
			slog.Info("torrent removed from queue", "id", id)
			return nil
		}
		return fmt.Errorf("manga not found: %s", id)
	}

	slog.Info("Dropping manga", "title", manga.Title(), "id", manga.Id(), "deleteFiles", deleteFiles)
	go func() {
		m.mangas.Delete(id)
		manga.Cancel()
		m.mu.Lock()
		m.downloading = nil
		m.mu.Unlock()

		if deleteFiles {
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
		err := m.Download(q.Id, q.BaseDir)
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

func (m *mangadexClientImpl) GetQueuedMangas() []config.QueueStat {
	return m.queue.Items()
}
