package webtoon

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"log/slog"
	"os"
	"path"
	"slices"
	"sync"
)

var c Client

func I() Client {
	return c
}

func newClient(c Config) Client {
	return &client{
		config:      c,
		webtoons:    utils.NewSafeMap[string, WebToon](),
		queue:       utils.NewQueue[payload.QueueStat](),
		downloading: nil,
		mu:          sync.Mutex{},
	}
}

type client struct {
	config      Config
	webtoons    *utils.SafeMap[string, WebToon]
	queue       utils.Queue[payload.QueueStat]
	downloading WebToon
	mu          sync.Mutex
}

func (c *client) Download(req payload.DownloadRequest) (WebToon, error) {
	if c.webtoons.Has(req.Id) {
		return nil, fmt.Errorf("webtoon alread exists: %s", req.Id)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.downloading != nil {
		c.queue.Enqueue(req.ToQueueStat())
		return nil, nil
	}

	log.Info("downloading webtoon",
		slog.String("id", req.Id),
		slog.String("into", req.BaseDir),
		slog.String("title?", req.TempTitle))
	wt := newWebToon(req, c)
	c.webtoons.Set(req.Id, wt)
	c.downloading = wt
	wt.WaitForInfoAndDownload()
	return wt, nil
}

func (c *client) RemoveDownload(request payload.StopRequest) error {
	wt, ok := c.webtoons.Get(request.Id)
	if !ok {
		ok = c.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == request.Id
		})
		if ok {
			log.Info("webtoon removed from queue", "id", request.Id)
			return nil
		}
		return fmt.Errorf("webtoon not found: %s", request.Id)
	}

	log.Info("dropping webtoon",
		slog.String("id", request.Id),
		slog.String("title", wt.Title()),
		slog.Bool("deleteFiles", request.DeleteFiles))
	go func() {
		c.webtoons.Delete(request.Id)
		wt.Cancel()
		c.mu.Lock()
		c.downloading = nil
		c.mu.Unlock()

		if request.DeleteFiles {
			go c.deleteFiles(wt)
		} else {
			go c.cleanup(wt)
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

func (c *client) deleteFiles(wt WebToon) {
	if !wt.Downloading() {
		return
	}

	dir := path.Join(c.GetBaseDir(), wt.GetDownloadDir())
	l := log.With(slog.String("dir", dir), slog.String("id", wt.Id()))

	skip := wt.GetPrevChapters()
	if len(skip) == 0 {
		l.Info("deleting directory")
		if err := os.RemoveAll(dir); err != nil {
			l.Error("error while deleting directory", slog.Any("error", err))
		}
		return
	}

	l.Info("deleting new entries in directory", slog.String("skipping", fmt.Sprintf("%+v", skip)))
	entries, err := os.ReadDir(dir)
	if err != nil {
		l.Error("error while reading directory", slog.Any("error", err))
		return
	}

	for _, entry := range entries {
		if slices.Contains(skip, entry.Name()) {
			l.Trace("skipping", slog.String("entry", entry.Name()))
			continue
		}

		log.Trace("deleting inner directory", slog.String("entry", entry.Name()))
		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error("error while deleting inner directory", slog.String("entry", entry.Name()), slog.Any("error", err))
		}
	}
}

func (c *client) cleanup(wt WebToon) {
	dir := path.Join(c.GetBaseDir(), wt.GetDownloadDir())
	l := log.With(slog.String("dir", dir), slog.String("id", wt.Id()))

	entries, err := os.ReadDir(dir)
	if err != nil {
		l.Error("error while reading directory", slog.Any("error", err))
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		err = utils.ZipFolder(path.Join(dir, entry.Name()), path.Join(dir, entry.Name()+".cbz"))
		if err != nil {
			l.Error("error while zipping file", slog.Any("error", err))
			continue
		}

		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error("error while deleting file", slog.Any("error", err))
			return
		}
	}
}

func (c *client) GetBaseDir() string {
	return config.OrDefault(c.config.GetRootDir(), "temp")
}

func (c *client) GetCurrentWebToon() WebToon {
	return c.downloading
}

func (c *client) GetQueuedWebToons() []payload.QueueStat {
	return c.queue.Items()
}

func (c *client) GetConfig() Config {
	return c.config
}
