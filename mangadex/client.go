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
		queue:       utils.NewQueue[string](),
		downloading: nil,
		mu:          sync.Mutex{},
	}
}

type mangadexClientImpl struct {
	dir         string
	mangas      *utils.SafeMap[string, Manga]
	queue       utils.Queue[string]
	downloading Manga
	mu          sync.Mutex
}

func (m *mangadexClientImpl) Download(id string, baseDir string) error {
	if m.mangas.Has(id) {
		return fmt.Errorf("manga already exists: %s", id)
	}

	manga := newManga(id, baseDir)
	m.mangas.Set(id, manga)

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.downloading != nil {
		m.queue.Enqueue(id)
	} else {
		m.downloading = manga
		manga.WaitForInfoAndDownload()
	}

	return nil
}

func (m *mangadexClientImpl) RemoveDownload(id string, deleteFiles bool) error {
	manga, ok := m.mangas.Get(id)
	if !ok {
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

	var manga Manga
	var ok bool
	for !ok && !m.queue.IsEmpty() {
		nextId, err := m.queue.Dequeue()
		if err != nil {
			slog.Debug("Error while dequeueing manga from queue", "error", err)
			return
		}
		manga, ok = m.mangas.Get(*nextId)
		if !ok {
			slog.Debug("manga not found", "id", nextId)
			return
		}
		manga.WaitForInfoAndDownload()
		m.mu.Lock()
		m.downloading = manga
		m.mu.Unlock()
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
