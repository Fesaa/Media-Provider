package mangadex

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"os"
	"path"
	"slices"
	"sync"
)

var m Client

func I() Client {
	return m
}

func newClient(c Config) Client {
	return &client{
		config:      c,
		mangas:      utils.NewSafeMap[string, Manga](),
		queue:       utils.NewQueue[payload.QueueStat](),
		downloading: nil,
		mu:          sync.Mutex{},
	}
}

type client struct {
	config      Config
	mangas      *utils.SafeMap[string, Manga]
	queue       utils.Queue[payload.QueueStat]
	downloading Manga
	mu          sync.Mutex
}

func (m *client) GetBaseDir() string {
	return config.OrDefault(m.config.GetRootDir(), "temp")
}

func (m *client) GetConfig() Config {
	return m.config
}

func (m *client) GetCurrentManga() Manga {
	return m.downloading
}

func (m *client) GetQueuedMangas() []payload.QueueStat {
	return m.queue.Items()
}

func (m *client) Download(req payload.DownloadRequest) (Manga, error) {
	if m.mangas.Has(req.Id) {
		return nil, fmt.Errorf("manga already exists: %s", req.Id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.downloading != nil {
		m.queue.Enqueue(req.ToQueueStat())
		return nil, nil
	}

	log.Info("downloading manga", "mangaId", req.Id, "into", req.BaseDir, "title?", req.TempTitle)
	manga := newManga(req, m.GetConfig(), m)
	m.mangas.Set(req.Id, manga)
	m.downloading = manga
	manga.WaitForInfoAndDownload()
	return manga, nil
}

func (m *client) RemoveDownload(req payload.StopRequest) error {
	manga, ok := m.mangas.Get(req.Id)
	if !ok {
		ok = m.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == req.Id
		})
		if ok {
			log.Info("manga removed from queue", "mangaId", req.Id)
			return nil
		}
		return fmt.Errorf("manga not found: %s", req.Id)
	}

	log.Info("Dropping manga", "mangaId", req.Id, "title", manga.Title(), "deleteFiles", req.DeleteFiles)
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

func (m *client) startNext() {
	if m.queue.IsEmpty() {
		return
	}

	added := false
	for !added && !m.queue.IsEmpty() {
		q, _ := m.queue.Dequeue()
		_, err := m.Download(q.ToDownloadRequest())
		if err != nil {
			log.Warn("error while adding manga from queue", "error", err)
			continue
		}
		added = true
	}
}

func (m *client) deleteFiles(manga Manga) {
	dir := path.Join(m.GetBaseDir(), manga.GetDownloadDir())
	skip := manga.GetPrevVolumes()
	if len(skip) == 0 {
		log.Info("deleting directory", "dir", dir, "mangaId", manga.Id())
		if err := os.RemoveAll(dir); err != nil {
			log.Error("error while deleting directory", "dir", dir, "mangaId", manga.Id(), "err", err)
		}
		return
	}

	log.Info("deleting new entries in directory", "dir", dir, "mangaId", manga.Id(), "skipping", fmt.Sprintf("%+v", skip))
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error("error while reading dir", "dir", dir, "mangaId", manga.Id(), "err", err)
		return
	}
	for _, entry := range entries {
		if slices.Contains(skip, entry.Name()) {
			log.Trace("skipping inner directory", "dir", dir, "mangaId", manga.Id(), "name", entry.Name())
			continue
		}

		log.Trace("deleting inner directory", "dir", dir, "mangaId", manga.Id(), "name", entry.Name())
		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			log.Error("error while deleting directory", "dir", dir, "mangaId", manga.Id(), "err", err)
		}
	}
}

func (m *client) cleanup(manga Manga) {
	dir := path.Join(m.GetBaseDir(), manga.GetBaseDir(), manga.Title())
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error("error while reading directory", "dir", dir, "mangaId", manga.Id(), "err", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		err = utils.ZipFolder(path.Join(dir, entry.Name()), path.Join(dir, entry.Name()+".cbz"))
		if err != nil {
			log.Error("error while zipping directory", "dir", dir, "mangaId", manga.Id(), "err", err)
			continue
		}

		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			log.Error("error while removing old directory", "dir", entry.Name(), "mangaId", manga.Id(), "err", err)
			return
		}
	}
}
