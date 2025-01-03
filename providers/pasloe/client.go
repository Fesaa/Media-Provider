package pasloe

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"net/http"
	"os"
	"path"
	"sync"
)

func newClient(c *config.Config, httpClient *http.Client, container *dig.Container, log zerolog.Logger) api.Client {
	return &client{
		config:   c,
		registry: newRegistry(httpClient, container),
		log:      log.With().Str("handler", "pasloe").Logger(),

		downloads:   utils.NewSafeMap[string, api.Downloadable](),
		queue:       utils.NewQueue[payload.QueueStat](),
		downloading: make([]api.Downloadable, 0),
		mu:          sync.Mutex{},
	}
}

type client struct {
	config   api.Config
	registry *registry
	log      zerolog.Logger

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

func (c *client) Download(req payload.DownloadRequest) error {
	if c.downloads.Has(req.Id) {
		return fmt.Errorf("manga already exists: %s", req.Id)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if utils.Any(c.downloading, func(downloadable api.Downloadable) bool {
		return downloadable.Provider() == req.Provider
	}) {
		c.queue.Enqueue(req.ToQueueStat())
		return nil
	}

	c.log.Debug().
		Str("id", req.Id).
		Str("into", req.BaseDir).
		Str("title?", req.TempTitle).
		Msg("downloading content")
	content, err := c.registry.Create(c, req)
	if err != nil {
		return err
	}
	c.downloads.Set(req.Id, content)
	c.downloading = append(c.downloading, content)
	content.WaitForInfoAndDownload()
	return nil
}

func (c *client) RemoveDownload(req payload.StopRequest) error {
	content, ok := c.downloads.Get(req.Id)
	if !ok {
		ok = c.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == req.Id
		})
		if ok {
			c.log.Info().Str("id", req.Id).Msg("content removed from queue")
			return nil
		}
		return fmt.Errorf("manga not found: %s", req.Id)
	}

	c.log.Info().
		Str("id", req.Id).
		Str("title", content.Title()).
		Bool("deleteFiles", req.DeleteFiles).
		Msg("removing content")
	go func() {
		c.downloads.Delete(req.Id)
		content.Cancel()
		c.mu.Lock()
		c.downloading = nil
		c.mu.Unlock()

		if req.DeleteFiles {
			go c.deleteFiles(content)
		} else {
			go c.cleanup(content)
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
		err := c.Download(q.ToDownloadRequest())
		if err != nil {
			c.log.Warn().Err(err).Msg("error while adding content from queue")
			continue
		}
		added = true
	}
}

func (c *client) deleteFiles(content api.Downloadable) {
	downloadDir := content.GetDownloadDir()
	if downloadDir == "" {
		c.log.Error().Msg("download dir is empty, not removing any files")
		return
	}
	dir := path.Join(c.GetBaseDir(), downloadDir)
	l := c.log.With().Str("dir", dir).Str("contentId", content.Id()).Logger()

	if len(content.GetOnDiskContent()) == 0 {
		l.Info().Msg("no existing content downloaded, removing entire directory")
		if err := os.RemoveAll(dir); err != nil {
			l.Error().Err(err).Msg("error while deleting directory")
		}
		return
	}

	for _, contentPath := range content.GetNewContent() {
		l.Trace().Str("path", contentPath).Msg("deleting new content dir")
		if err := os.RemoveAll(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while new content dir")
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		l.Error().Err(err).Str("dir", dir).Msg("error while reading dir, unable to remove empty dirs")
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		innerEntries, err := os.ReadDir(path.Join(dir, entry.Name()))
		if err != nil {
			l.Error().Err(err).Str("dir", dir).Str("name", entry.Name()).
				Msg("error while reading dir, will not remove")
			continue
		}

		if len(innerEntries) > 0 {
			l.Trace().Str("dir", dir).Str("name", entry.Name()).
				Msg("Dir has content, not removing any files")
			continue
		}

		l.Trace().Str("dir", dir).Str("name", entry.Name()).
			Msg("Dir has content, removing entire directory")
		if err := os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error().Err(err).Str("name", entry.Name()).Msg("error while new content dir")
		}
	}
}

func (c *client) cleanup(content api.Downloadable) {
	l := c.log.With().Str("contentId", content.Id()).Logger()
	for _, contentPath := range content.GetNewContent() {
		l.Debug().Str("path", contentPath).Msg("Zipping file")
		err := utils.ZipFolder(contentPath, contentPath+".cbz")
		if err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while zipping dir")
			continue
		}

		if err = os.RemoveAll(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while deleting file")
			return
		}
	}
}
