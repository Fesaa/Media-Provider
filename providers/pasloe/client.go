package pasloe

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"net/http"
	"os"
	"path"
	"time"
)

func New(c *config.Config, httpClient *http.Client, container *dig.Container, log zerolog.Logger,
	ioService services.IOService, signalR services.SignalRService, notify services.NotificationService,
	preferences models.Preferences, transLoco services.TranslocoService,
) api.Client {
	return &client{
		config:    c,
		registry:  newRegistry(httpClient, container),
		log:       log.With().Str("handler", "pasloe").Logger(),
		io:        ioService,
		signalR:   signalR,
		notify:    notify,
		pref:      preferences,
		transLoco: transLoco,

		content: utils.NewSafeMap[string, api.Downloadable](),
	}
}

type client struct {
	config    api.Config
	registry  *registry
	log       zerolog.Logger
	io        services.IOService
	signalR   services.SignalRService
	notify    services.NotificationService
	transLoco services.TranslocoService
	pref      models.Preferences

	content utils.SafeMap[string, api.Downloadable]
}

func (c *client) Content(id string) services.Content {
	content, ok := c.content.Get(id)
	if !ok {
		return nil
	}
	return content
}

func (c *client) Download(req payload.DownloadRequest) error {
	if c.content.Has(req.Id) {
		return c.wrapError(services.ErrContentAlreadyExists)
	}

	content, err := c.registry.Create(c, req)
	if err != nil {
		return c.wrapError(err)
	}

	c.content.Set(content.Id(), content)
	c.signalR.AddContent(content.GetInfo())

	if !c.CanStart(content.Provider()) {
		return nil
	}

	go func() {
		content.StartLoadInfo()

		c.loadAllInfo(content.Provider())

		// We are certain it's fine to start, as CanStart was true (i.e. the provider has nothing making requests)
		if content.State() == payload.ContentStateReady {
			c.log.Debug().
				Str("id", content.Id()).
				Str("into", content.GetBaseDir()).
				Str("title", content.Title()).
				Msg("downloading content")
			content.StartDownload()
		} else if content.State() == payload.ContentStateWaiting {
			c.log.Debug().
				Str("id", content.Id()).
				Str("title", content.Title()).
				Msg("Content cannot be downloaded yet, checking if an other can start")
			c.startNext(content.Provider())
		}
	}()
	return nil
}

func (c *client) RemoveDownload(req payload.StopRequest) error {
	content, ok := c.content.Get(req.Id)
	if !ok {
		return c.wrapError(services.ErrContentNotFound)
	}

	// Delete early to ensure no follow-up requests can be made to it
	c.content.Delete(content.Id())
	if content.State() != payload.ContentStateDownloading {
		// The removed content has not written anything to disk yet. Nothing to clean up,
		// and no need to start the next content.
		c.log.Info().
			Str("id", req.Id).
			Str("title", content.Title()).
			Msg("content won't be downloaded")

		// Cancelling loading information
		if content.State() == payload.ContentStateLoading {
			go content.Cancel()
		}

		c.signalR.DeleteContent(content.Id())

		return nil
	}

	c.log.Info().
		Str("id", req.Id).
		Str("title", content.Title()).
		Bool("deleteFiles", req.DeleteFiles).
		Msg("removing content")
	go func() {
		content.Cancel()
		c.signalR.StateUpdate(content.Id(), payload.ContentStateCleanup)

		if req.DeleteFiles {
			go c.deleteFiles(content)
			c.startNext(content.Provider())
			return
		}

		alwaysLog := utils.TryCatch(c.pref.Get, func(p *models.Preference) bool {
			return p.LogEmptyDownloads
		}, false, func(err error) {
			c.log.Error().Err(err).Msg("failed to retrieve preferences, falling back to default behaviour")
		})

		if len(content.GetNewContent()) > 0 || alwaysLog {
			text := c.transLoco.GetTranslation("download-finished", content.Title(), len(content.GetNewContent()))
			if len(content.GetToRemoveContent()) > 0 {
				text += c.transLoco.GetTranslation("re-downloads", len(content.GetToRemoveContent()))
			}
			c.notifier(content.Request()).Notify(models.Notification{
				Title:   c.transLoco.GetTranslation("download-finished-title"),
				Summary: utils.Shorten(text, services.SummarySize),
				Body:    text,
				Colour:  models.Green,
				Group:   models.GroupContent,
			})
		}

		go c.cleanup(content)
		c.startNext(content.Provider())
	}()
	return nil
}

func (c *client) GetCurrentDownloads() []api.Downloadable {
	return c.content.Values()
}

func (c *client) GetBaseDir() string {
	return config.OrDefault(c.config.GetRootDir(), "temp")
}

func (c *client) GetConfig() api.Config {
	return c.config
}

func (c *client) CanStart(provider models.Provider) bool {
	providerBusy := c.content.Any(func(k string, d api.Downloadable) bool {
		return d.Provider() == provider &&
			d.State() > payload.ContentStateQueued &&
			d.State() != payload.ContentStateWaiting &&
			d.State() < payload.ContentStateCleanup
	})

	return !providerBusy
}

func (c *client) notifier(req payload.DownloadRequest) services.Notifier {
	if req.IsSubscription {
		return c.notify
	}
	return c.signalR
}

func (c *client) loadAllInfo(provider models.Provider) {
	nextQueue := func(id string, d api.Downloadable) bool {
		return d.Provider() == provider && d.State() == payload.ContentStateQueued
	}

	for {
		next, ok := c.content.Find(nextQueue)
		if !ok {
			break
		}

		(*next).StartLoadInfo()
	}
}

func (c *client) startNext(provider models.Provider) {
	c.loadAllInfo(provider)

	inext, ok := c.content.Find(func(k string, d api.Downloadable) bool {
		return d.Provider() == provider && d.State() == payload.ContentStateReady
	})
	if !ok {
		return
	}

	next := *inext
	c.log.Debug().
		Str("id", next.Id()).
		Str("into", next.GetBaseDir()).
		Str("title", next.Title()).
		Msg("downloading content")
	c.signalR.Notify(models.Notification{
		Title:   "Now starting",
		Summary: next.Title(),
		Colour:  models.Blue,
		Group:   models.GroupContent,
	})
	next.StartDownload()
}

// TODO: Rewrite
//
//nolint:funlen
func (c *client) deleteFiles(content api.Downloadable) {
	defer c.signalR.DeleteContent(content.Id())

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
			c.notifyCleanUpError(content, err)
		}
		return
	}

	var cleanupErrs []error
	defer func() {
		if len(cleanupErrs) > 0 {
			c.notifyCleanUpError(content, cleanupErrs...)
		}
	}()

	start := time.Now()
	for _, contentPath := range content.GetNewContent() {
		l.Trace().Str("path", contentPath).Msg("deleting new content dir")
		if err := os.RemoveAll(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while removing new content dir")
			cleanupErrs = append(cleanupErrs, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		l.Error().Err(err).Str("dir", dir).Msg("error while reading dir, unable to remove empty dirs")
		cleanupErrs = append(cleanupErrs, err)
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
			cleanupErrs = append(cleanupErrs, err)
			continue
		}

		if len(innerEntries) > 0 {
			l.Trace().Str("dir", dir).Str("name", entry.Name()).
				Msg("Dir has content, not removing any files")
			continue
		}

		l.Trace().Str("dir", dir).Str("name", entry.Name()).
			Msg("Dir has no content, removing entire directory")
		if err := os.Remove(path.Join(dir, entry.Name())); err != nil {
			l.Error().Err(err).Str("name", entry.Name()).Msg("error while new content dir")
			cleanupErrs = append(cleanupErrs, err)
		}
	}
	l.Debug().Dur("elapsed", time.Since(start)).Msg("finished removing newly downloaded files")
}

// TODO: Rewrite
func (c *client) cleanup(content api.Downloadable) {
	defer c.signalR.DeleteContent(content.Id())

	l := c.log.With().Str("contentId", content.Id()).Logger()
	newContent := content.GetNewContent()
	if len(newContent) == 0 {
		return
	}

	var cleanupErrs []error

	start := time.Now()
	for _, contentPath := range content.GetToRemoveContent() {
		l.Trace().Str("name", contentPath).Msg("removing old content")
		if err := os.Remove(contentPath); err != nil {
			l.Error().Err(err).Str("name", contentPath).Msg("error while removing old content")
			cleanupErrs = append(cleanupErrs, err)
		}
	}
	l.Debug().Dur("elapsed", time.Since(start)).Int("size", len(content.GetToRemoveContent())).
		Msg("finished removing replaced downloaded content")

	start = time.Now()
	for _, contentPath := range newContent {
		l.Trace().Str("path", contentPath).Msg("Zipping file")
		err := c.io.ZipToCbz(contentPath)
		if err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while zipping dir")
			cleanupErrs = append(cleanupErrs, err)
			continue
		}

		if err = os.RemoveAll(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while deleting new content directory")
			cleanupErrs = append(cleanupErrs, err)
			continue
		}
	}

	l.Debug().Dur("elapsed", time.Since(start)).Int("size", len(newContent)).
		Msg("finished zipping newly downloaded content")

	if len(cleanupErrs) > 0 {
		l.Error().Errs("errors", cleanupErrs).Msg("errors encountered during cleanup")
		c.notifyCleanUpError(content, cleanupErrs...)
	}
}

func (c *client) notifyCleanUpError(content api.Downloadable, cleanupErrs ...error) {
	joinedErr := errors.Join(cleanupErrs...)
	if joinedErr == nil {
		return
	}
	c.notify.NotifyContent(
		c.transLoco.GetTranslation("cleanup-errors-title"),
		c.transLoco.GetTranslation("cleanup-errors-summary", content.Title()),
		joinedErr.Error(),
		models.Red)
}

func (c *client) wrapError(err error) error {
	return fmt.Errorf("pasloe client error: %w", err)
}
