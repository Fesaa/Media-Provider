package pasloe

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/utils"
	"os"
	"path"
	"slices"
	"sync"
)

func newClient(c api.Config) api.Client {
	return &client{
		config:   c,
		registry: newRegistry(),

		downloads:   utils.NewSafeMap[string, api.Downloadable](),
		queue:       utils.NewQueue[payload.QueueStat](),
		downloading: make([]api.Downloadable, 0),
		mu:          sync.Mutex{},
	}
}

type client struct {
	config   api.Config
	registry *registry

	downloads   *utils.SafeMap[string, api.Downloadable]
	queue       utils.Queue[payload.QueueStat]
	downloading []api.Downloadable
	mu          sync.Mutex
}

func (c *client) GetBaseDir() string {
	return config.OrDefault(c.config.GetRootDir(), "temp")
}

func (c *client) GetConfig() api.Config {
	return c.config
}

func (c *client) GetCurrentDownloads() []api.Downloadable {
	return c.downloading
}

func (c *client) GetQueuedDownloads() []payload.QueueStat {
	return c.queue.Items()
}

func (c *client) Download(req payload.DownloadRequest) (api.Downloadable, error) {
	if c.downloads.Has(req.Id) {
		return nil, fmt.Errorf("manga already exists: %s", req.Id)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if utils.Any(c.downloading, func(downloadable api.Downloadable) bool {
		return downloadable.Provider() == req.Provider
	}) {
		c.queue.Enqueue(req.ToQueueStat())
		return nil, nil
	}

	log.Info("downloading content", "id", req.Id, "into", req.BaseDir, "title?", req.TempTitle)
	content, err := c.registry.Create(c, req)
	if err != nil {
		return nil, err
	}
	c.downloads.Set(req.Id, content)
	c.downloading = append(c.downloading, content)
	content.WaitForInfoAndDownload()
	return content, nil
}

func (c *client) RemoveDownload(req payload.StopRequest) error {
	manga, ok := c.downloads.Get(req.Id)
	if !ok {
		ok = c.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == req.Id
		})
		if ok {
			log.Info("manga removed from queue", "mangaId", req.Id)
			return nil
		}
		return fmt.Errorf("manga not found: %s", req.Id)
	}

	log.Info("dropping manga", "mangaId", req.Id, "title", manga.Title(), "deleteFiles", req.DeleteFiles)
	go func() {
		c.downloads.Delete(req.Id)
		manga.Cancel()
		c.mu.Lock()
		c.downloading = nil
		c.mu.Unlock()

		if req.DeleteFiles {
			go c.deleteFiles(manga)
		} else {
			go c.cleanup(manga)
		}
		c.startNext()
	}()
	return nil
}

func (c *client) startNext() {
	if c.queue.IsEmpty() {
		return
	}

	added := false
	for !added && !c.queue.IsEmpty() {
		q, _ := c.queue.Dequeue()
		_, err := c.Download(q.ToDownloadRequest())
		if err != nil {
			log.Warn("error while adding manga from queue", "error", err)
			continue
		}
		added = true
	}
}

func (c *client) deleteFiles(content api.Downloadable) {
	downloadDir := content.GetDownloadDir()
	if downloadDir == "" {
		log.Error("download dir is empty, not removing any files")
		return
	}
	dir := path.Join(c.GetBaseDir(), downloadDir)
	skip := content.GetOnDiskContent()

	l := log.With("dir", dir, "contentId", content.Id())

	if len(skip) == 0 {
		l.Info("deleting directory")
		if err := os.RemoveAll(dir); err != nil {
			l.Error("error while deleting directory", "err", err)
		}
		return
	}

	l.Info("deleting new entries in directory", "skipping", fmt.Sprintf("%+v", skip))
	entries, err := os.ReadDir(dir)
	if err != nil {
		l.Error("error while reading dir", "err", err)
		return
	}
	for _, entry := range entries {
		if slices.Contains(skip, entry.Name()) {
			l.Trace("skipping inner directory", "name", entry.Name())
			continue
		}

		l.Trace("deleting inner directory", "name", entry.Name())
		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error("error while deleting directory", "err", err)
		}
	}
}

func (c *client) cleanup(content api.Downloadable) {
	dir := path.Join(c.GetBaseDir(), content.GetBaseDir(), content.Title())
	entries, err := os.ReadDir(dir)

	l := log.With("dir", dir, "contentId", content.Id())

	if err != nil {
		l.Error("error while reading directory", "err", err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		err = utils.ZipFolder(path.Join(dir, entry.Name()), path.Join(dir, entry.Name()+".cbz"))
		if err != nil {
			l.Error("error while zipping directory", "err", err)
			continue
		}

		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error("error while removing old directory", "dir", entry.Name(), "err", err)
			return
		}
	}
}
