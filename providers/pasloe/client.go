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
	"slices"
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

	c.log.Info().
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
	skip := content.GetOnDiskContent()

	l := c.log.With().Str("dir", dir).Str("contentId", content.Id()).Logger()

	if len(skip) == 0 {
		l.Info().Msg("deleting directory")
		if err := os.RemoveAll(dir); err != nil {
			l.Error().Err(err).Msg("error while deleting directory")
		}
		return
	}

	l.Info().Str("skipping", fmt.Sprintf("%+v", skip)).Msg("deleting new entries in directory")
	entries, err := os.ReadDir(dir)
	if err != nil {
		l.Error().Err(err).Msg("error while reading directory")
		return
	}
	for _, entry := range entries {
		if slices.Contains(skip, entry.Name()) {
			l.Trace().Str("name", entry.Name()).Msg("skipping inner dir")
			continue
		}

		l.Trace().Str("name", entry.Name()).Msg("deleting inner directory")
		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error().Err(err).Str("name", entry.Name()).Msg("error while deleting directory")
		}
	}
}

func (c *client) cleanup(content api.Downloadable) {
	dir := path.Join(c.GetBaseDir(), content.GetBaseDir(), content.Title())
	entries, err := os.ReadDir(dir)

	l := c.log.With().Str("dir", dir).Str("contentId", content.Id()).Logger()

	if err != nil {
		l.Error().Err(err).Msg("error while reading directory")
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		err = utils.ZipFolder(path.Join(dir, entry.Name()), path.Join(dir, entry.Name()+".cbz"))
		if err != nil {
			l.Error().Err(err).Str("name", entry.Name()).Msg("error while zipping dir")
			continue
		}

		if err = os.RemoveAll(path.Join(dir, entry.Name())); err != nil {
			l.Error().Err(err).Str("name", entry.Name()).Msg("error while deleting file")
			return
		}
	}
}
