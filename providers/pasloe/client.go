package pasloe

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
)

func New(s services.SettingsService, container *dig.Container, log zerolog.Logger,
	dirService services.DirectoryService, signalR services.SignalRService, notify services.NotificationService,
	preferences models.Preferences, transLoco services.TranslocoService, fs afero.Afero,
) (core.Client, error) {
	settings, err := s.GetSettingsDto()
	if err != nil {
		return nil, err
	}

	clientCtx, cancel := context.WithCancel(context.Background())

	return &client{
		registry:   newRegistry(container),
		log:        log.With().Str("handler", "pasloe").Logger(),
		dirService: dirService,
		signalR:    signalR,
		notify:     notify,
		pref:       preferences,
		transLoco:  transLoco,
		fs:         fs,

		content:        utils.NewSafeMap[string, core.Downloadable](),
		providerQueues: utils.NewSafeMap[models.Provider, *ProviderQueue](),
		rootDir:        settings.RootDir,
		ctx:            clientCtx,
		cancel:         cancel,
		deletionWg:     &sync.WaitGroup{},
	}, nil
}

type client struct {
	registry Registry
	log      zerolog.Logger

	dirService services.DirectoryService
	signalR    services.SignalRService
	notify     services.NotificationService
	transLoco  services.TranslocoService
	pref       models.Preferences
	fs         afero.Afero

	rootDir string

	content        utils.SafeMap[string, core.Downloadable]
	providerQueues utils.SafeMap[models.Provider, *ProviderQueue]
	mu             sync.RWMutex

	ctx        context.Context
	cancel     context.CancelFunc
	deletionWg *sync.WaitGroup
}

// getOrCreateProviderQueue returns the existing queue for the provider, or creates a new one if none found
// this defers the worker threads for each provider until they're used
func (c *client) getOrCreateProviderQueue(provider models.Provider) *ProviderQueue {
	if pq, ok := c.providerQueues.Get(provider); ok {
		return pq
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring lock
	if pq, exists := c.providerQueues.Get(provider); exists {
		return pq
	}

	pq := NewProviderQueue(provider, c.ctx, c, c.log)
	c.providerQueues.Set(provider, pq)
	return pq

}

// Download queues content to be downloaded
func (c *client) Download(req payload.DownloadRequest) error {
	if c.content.Has(req.Id) {
		return c.wrapError(services.ErrContentAlreadyExists)
	}

	content, err := c.registry.Create(c, req)
	if err != nil {
		return c.wrapError(err)
	}

	c.content.Set(content.Id(), content)
	c.signalR.AddContent(content.Request().OwnerId, content.GetInfo())

	pq := c.getOrCreateProviderQueue(content.Provider())

	if err = pq.AddToLoadingQueue(content); err != nil {
		c.content.Delete(content.Id())
		c.signalR.DeleteContent(content.Id())

		return c.wrapError(err)
	}

	return nil
}

// MoveToDownloadQueue forcefully move content with the given id to the download queue
func (c *client) MoveToDownloadQueue(id string) error {
	content, ok := c.content.Get(id)
	if !ok {
		return services.ErrContentNotFound
	}

	pq := c.getOrCreateProviderQueue(content.Provider())
	return pq.AddToDownloadQueue(content)
}

// RemoveDownload stop content from being downloaded, optionally remove newly downloaded files. Otherwise, clean up
func (c *client) RemoveDownload(req payload.StopRequest) error {
	content, ok := c.content.Get(req.Id)
	if !ok {
		return c.wrapError(services.ErrContentNotFound)
	}

	c.content.Delete(content.Id())

	c.log.Info().
		Str("id", req.Id).
		Str("title", content.Title()).
		Bool("deleteFiles", req.DeleteFiles).
		Msg("removing content")

	c.deletionWg.Add(1)
	go func() {
		defer c.deletionWg.Done()

		content.Cancel()
		c.signalR.StateUpdate(content.Request().OwnerId, content.Id(), payload.ContentStateCleanup)

		if req.DeleteFiles {
			c.deleteFiles(content)
		} else {
			c.logContentCompletion(content)
			c.cleanup(content)
		}

		c.signalR.DeleteContent(content.Id())
	}()

	return nil
}

func (c *client) Content(id string) services.Content {
	content, ok := c.content.Get(id)
	if !ok {
		return nil
	}
	return content
}

// Shutdown gracefully shuts down the client
func (c *client) Shutdown() error {
	c.log.Debug().Msg("pasloe shutting down")

	c.cancel()

	c.providerQueues.ForEach(func(k models.Provider, v *ProviderQueue) {
		v.Shutdown()
	})

	c.content.ForEach(func(k string, v core.Downloadable) {
		if err := c.RemoveDownload(payload.StopRequest{Id: k, DeleteFiles: true, Provider: v.Provider()}); err != nil {
			c.log.Warn().Err(err).Msg("failed to remove download")
		}
	})

	c.log.Debug().Msg("Stop requests send out, waiting for all deletion to finish")
	utils.WaitFor(c.deletionWg, time.Second*45)

	c.log.Debug().Msg("pasloe shutdown complete")

	return nil
}

func (c *client) alwaysLog() bool {
	p, err := c.pref.Get()
	if err != nil {
		c.log.Error().Err(err).Msg("failed to retrieve preferences, falling back to default behaviour")
		return false
	}

	return p.LogEmptyDownloads
}

func (c *client) logContentCompletion(content core.Downloadable) {
	alwaysLog := c.alwaysLog()

	if len(content.GetNewContent()) == 0 && !alwaysLog {
		return
	}

	var summary, body string

	di := content.DisplayInformation()

	summary = c.transLoco.GetTranslation("download-finished", content.GetInfo().RefUrl, di.Name, len(content.GetNewContent()))
	if len(content.GetToRemoveContent()) > 0 {
		summary += c.transLoco.GetTranslation("re-downloads", len(content.GetToRemoveContent()))
	}

	if content.FailedDownloads() > 0 {
		summary += c.transLoco.GetTranslation("failed-downloads", content.FailedDownloads())
	}

	body = summary
	for _, newContent := range content.GetNewContentNamed() {
		body += c.transLoco.GetTranslation("content-line", path.Base(newContent))
	}

	c.notifier(content.Request()).Notify(models.Notification{
		Title:   c.transLoco.GetTranslation("download-finished-title"),
		Summary: summary,
		Body:    body,
		Colour:  models.Secondary,
		Group:   models.GroupContent,
	})
}

func (c *client) GetCurrentDownloads() []core.Downloadable {
	return c.content.Values()
}

func (c *client) GetBaseDir() string {
	return utils.OrElse(c.rootDir, "temp")
}

func (c *client) notifier(req payload.DownloadRequest) services.Notifier {
	if req.IsSubscription {
		return c.notify
	}
	return c.signalR
}

func (c *client) deleteFiles(content core.Downloadable) {
	defer c.signalR.DeleteContent(content.Id())

	downloadDir := strings.TrimSpace(content.GetDownloadDir())
	if downloadDir == "" {
		c.log.Error().Msg("download dir is empty, not removing any files")
		return
	}

	start := time.Now()
	dir := path.Join(c.GetBaseDir(), downloadDir)

	l := c.log.With().Str("dir", dir).Str("contentId", content.Id()).Logger()

	// Skip if the directory is not found. We're not logging the error as it's safe to assume your logs
	// will be completely spammed if an error would occur here
	if ok, err := c.fs.DirExists(dir); ok && err == nil {

		cleanupErrs := c.deleteNewContent(content, l)
		cleanupErrs = append(cleanupErrs, c.deleteEmptyDirectories(dir, l)...)

		c.notifyCleanUpError(content, cleanupErrs...)
	}

	l.Debug().Dur("elapsed", time.Since(start)).Msg("finished removing newly downloaded files")
}

func (c *client) deleteNewContent(content core.Downloadable, l zerolog.Logger) (cleanupErrs []error) {
	for _, contentPath := range content.GetNewContent() {
		l.Trace().Str("path", contentPath).Msg("deleting new content dir")

		if err := c.fs.RemoveAll(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while removing new content dir")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error removing new content dir %s: %w", contentPath, err))
		}
	}

	return cleanupErrs
}

func (c *client) deleteEmptyDirectories(dir string, l zerolog.Logger) (cleanupErrs []error) {
	entries, err := c.fs.ReadDir(dir)
	if err != nil {
		l.Error().Err(err).Str("dir", dir).Msg("error while reading dir, unable to remove empty dirs")
		return append(cleanupErrs, fmt.Errorf("failed to read directory %s: %w", dir, err))
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		innerEntries, err := c.fs.ReadDir(path.Join(dir, entry.Name()))
		if err != nil {
			l.Error().Err(err).Str("dir", dir).Str("name", entry.Name()).
				Msg("error while reading dir, will not remove")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error reading dir %s: %w", entry.Name(), err))
			continue
		}

		if len(innerEntries) > 0 {
			l.Trace().Str("dir", dir).Str("name", entry.Name()).
				Msg("Dir has content, not removing any files")
			continue
		}

		l.Trace().Str("dir", dir).Str("name", entry.Name()).
			Msg("Dir has no content, removing entire directory")

		if err := c.fs.Remove(path.Join(dir, entry.Name())); err != nil {
			l.Error().Err(err).Str("name", entry.Name()).Msg("error while new content dir")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error removing dir %s: %w", entry.Name(), err))
		}
	}

	entries, err = c.fs.ReadDir(dir)
	if err != nil {
		l.Error().Err(err).Str("dir", dir).Msg("error while reading dir, unable to remove if empty")
		return append(cleanupErrs, fmt.Errorf("error reading dir %s: %w", dir, err))
	}

	if len(entries) == 0 {
		if err = c.fs.Remove(dir); err != nil {
			l.Error().Err(err).Str("dir", dir).Msg("error while removing empty series dir")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error removing dir %s: %w", dir, err))
		}
	}

	return cleanupErrs
}

func (c *client) cleanup(content core.Downloadable) {
	defer c.signalR.DeleteContent(content.Id())

	l := c.log.With().Str("contentId", content.Id()).Logger()
	newContent := content.GetNewContent()
	if len(newContent) == 0 {
		return
	}

	start := time.Now()

	cleanupErrs := c.removeOldContent(content, l)
	cleanupErrs = append(cleanupErrs, c.zipAndRemoveNewContent(newContent, l)...)

	if len(cleanupErrs) > 0 {
		l.Error().Errs("errors", cleanupErrs).Msg("errors encountered during cleanup")
		c.notifyCleanUpError(content, cleanupErrs...)
	}

	l.Debug().Dur("elapsed", time.Since(start)).Msg("finished cleanup")
}

func (c *client) removeOldContent(content core.Downloadable, l zerolog.Logger) (cleanupErrs []error) {
	start := time.Now()

	for _, contentPath := range content.GetToRemoveContent() {
		l.Trace().Str("name", contentPath).Msg("removing old content")

		if err := c.fs.Remove(contentPath); err != nil {
			l.Error().Err(err).Str("name", contentPath).Msg("error while removing old content")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error while removing old content: %w", err))
		}
	}

	l.Debug().Dur("elapsed", time.Since(start)).Int("size", len(content.GetToRemoveContent())).
		Msg("finished removing replaced downloaded content")
	return cleanupErrs
}

func (c *client) zipAndRemoveNewContent(newContent []string, l zerolog.Logger) (cleanupErrs []error) {
	start := time.Now()

	for _, contentPath := range newContent {
		l.Trace().Str("path", contentPath).Msg("Zipping file")

		if err := c.dirService.ZipToCbz(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while zipping dir")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error while zipping dir %s: %w", contentPath, err))
			continue
		}

		if err := c.fs.RemoveAll(contentPath); err != nil {
			l.Error().Err(err).Str("path", contentPath).Msg("error while deleting new content directory")
			cleanupErrs = append(cleanupErrs, fmt.Errorf("error while deleting new content directory %s: %w", contentPath, err))
			continue
		}
	}

	l.Debug().Dur("elapsed", time.Since(start)).Int("size", len(newContent)).
		Msg("finished zipping newly downloaded content")
	return cleanupErrs
}

func (c *client) notifyCleanUpError(content core.Downloadable, cleanupErrs ...error) {
	if len(cleanupErrs) == 0 {
		return
	}

	joinedErr := errors.Join(cleanupErrs...)
	if joinedErr == nil {
		return
	}
	c.notify.NotifyContent(
		c.transLoco.GetTranslation("cleanup-errors-title"),
		c.transLoco.GetTranslation("cleanup-errors-summary", content.Title()),
		joinedErr.Error(),
		models.Error)
}

func (c *client) wrapError(err error) error {
	return fmt.Errorf("pasloe client error: %w", err)
}
