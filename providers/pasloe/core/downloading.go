package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"slices"
	"sync"
	"time"
)

func (c *Core[T]) abortDownload(reason error) {
	if errors.Is(reason, context.Canceled) {
		return
	}

	c.Log.Error().Err(reason).Msg("error while downloading content; cleaning up")
	req := payload.StopRequest{
		Provider:    c.Req.Provider,
		Id:          c.Id(),
		DeleteFiles: true,
	}
	if err := c.Client.RemoveDownload(req); err != nil {
		c.Log.Error().Err(err).Msg("error while cleaning up")
	}
	c.Notifier.Notify(models.Notification{
		Title:   "Failed download",
		Summary: fmt.Sprintf("%s failed to download", c.infoProvider.Title()),
		Body:    fmt.Sprintf("Download failed for %s, because %v", c.infoProvider.Title(), reason),
		Colour:  models.Red,
		Group:   models.GroupError,
	})
}

func (c *Core[T]) filterContentByUserSelection() {
	if len(c.ToDownloadUserSelected) > 0 {
		currentSize := len(c.ToDownload)
		c.ToDownload = utils.Filter(c.ToDownload, func(t T) bool {
			return slices.Contains(c.ToDownloadUserSelected, t.GetId())
		})
		c.Log.Debug().Int("size", currentSize).Int("newSize", len(c.ToDownload)).
			Msg("content further filtered after user has made a selection in the UI")

		if len(c.ToRemoveContent) > 0 {
			paths := utils.Map(c.ToDownload, func(t T) string {
				return c.infoProvider.ContentPath(t) + ".cbz"
			})
			c.ToRemoveContent = utils.Filter(c.ToRemoveContent, func(s string) bool {
				return slices.Contains(paths, s)
			})
		}
	}
}

func (c *Core[T]) processDownloads(ctx context.Context) error {
	for _, content := range c.ToDownload {
		select {
		case <-ctx.Done():
			c.Wg.Wait()
			return errors.New("download cancelled")
		default:
			c.Wg.Add(1)
			err := c.downloadContent(ctx, content)
			c.Wg.Done()
			if err != nil {
				c.abortDownload(err)
				c.Wg.Wait()
				return err
			}
		}
		c.UpdateProgress()
	}
	return nil
}

func (c *Core[T]) cleanupAfterDownload() {
	c.Wg.Wait()
	req := payload.StopRequest{
		Provider:    c.Req.Provider,
		Id:          c.Id(),
		DeleteFiles: false,
	}
	if err := c.Client.RemoveDownload(req); err != nil {
		c.Log.Error().Err(err).Msg("error while cleaning up files")
	}
}

func (c *Core[T]) startDownload() {
	// Overwrite cancel, as we're doing something else
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	data := c.infoProvider.All()
	c.Log.Trace().Int("size", len(data)).Msg("downloading content")
	c.Wg = &sync.WaitGroup{}

	c.filterContentByUserSelection()

	c.Log.Info().
		Int("all", len(data)).
		Int("toDownload", len(c.ToDownload)).
		Int("reDownloads", len(c.ToRemoveContent)).
		Str("into", c.GetDownloadDir()).
		Msg("downloading content")

	if err := c.processDownloads(ctx); err != nil {
		c.Log.Trace().Err(err).Msg("download failed")
		return
	}

	c.cleanupAfterDownload()
}

type downloadUrl struct {
	idx int
	url string
}

func (c *Core[T]) downloadContent(ctx context.Context, t T) error {
	l := c.infoProvider.ContentLogger(t)
	l.Trace().Msg("downloading content")

	contentPath := c.infoProvider.ContentPath(t)
	if err := c.fs.MkdirAll(contentPath, 0755); err != nil {
		return err
	}
	c.HasDownloaded = append(c.HasDownloaded, contentPath)

	urls, err := c.infoProvider.ContentUrls(ctx, t)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		l.Warn().Msg("content has no downloadable urls? Unexpected? Report this!")
		return nil
	}

	if err = c.infoProvider.WriteContentMetaData(t); err != nil {
		c.Log.Warn().Err(err).Msg("error writing meta data")
	}

	l.Debug().Int("size", len(urls)).Msg("downloading images")

	urlCh := make(chan downloadUrl, c.maxImages)
	errCh := make(chan error, 1)
	defer close(errCh)

	wg := &sync.WaitGroup{}
	innerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.produceURLs(ctx, innerCtx, urls, urlCh)

	c.startProgressUpdater(ctx, innerCtx)

	for range c.maxImages {
		wg.Add(1)
		go c.channelConsumer(innerCtx, cancel, ctx, t, len(urls), l, urlCh, errCh, wg)
	}

	wg.Wait()

	select {
	case err = <-errCh:
		return err
	default:
	}

	if len(urls) < 5 {
		time.Sleep(1 * time.Second)
	}

	c.ContentDownloaded++
	return nil
}

func (c *Core[T]) produceURLs(ctx context.Context, innerCtx context.Context, urls []string, urlCh chan<- downloadUrl) {
	go func() {
		defer close(urlCh)
		for i, url := range urls {
			select {
			case <-ctx.Done():
				return
			case <-innerCtx.Done():
				return
			case urlCh <- downloadUrl{url: url, idx: i + 1}:
			}
		}
	}()
}

func (c *Core[T]) startProgressUpdater(ctx context.Context, innerCtx context.Context) {
	go func() {
		for range time.Tick(2 * time.Second) {
			select {
			case <-innerCtx.Done():
				return
			case <-ctx.Done():
				return
			default:
				c.UpdateProgress()
			}
		}
	}()
}

func (c *Core[T]) channelConsumer(
	innerCtx context.Context,
	cancel context.CancelFunc,
	ctx context.Context,
	t T,
	size int,
	l zerolog.Logger,
	urlCh chan downloadUrl,
	errCh chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	failedCh := make(chan downloadUrl, size)

	c.processInitialDownloads(innerCtx, ctx, t, l, urlCh, failedCh)

	select {
	case <-innerCtx.Done():
	case <-ctx.Done():
		return
	default:
		c.processFailedDownloads(innerCtx, ctx, t, l, failedCh, errCh, cancel)
	}
}

func (c *Core[T]) processInitialDownloads(
	innerCtx context.Context,
	ctx context.Context,
	t T,
	l zerolog.Logger,
	urlCh chan downloadUrl,
	failedCh chan downloadUrl,
) {
	for urlData := range urlCh {
		select {
		case <-innerCtx.Done():
			return
		case <-ctx.Done():
			return
		default:
			c.downloadURL(innerCtx, ctx, t, l, urlData, failedCh)
		}
	}
	close(failedCh)
}

func (c *Core[T]) downloadURL(
	innerCtx context.Context,
	ctx context.Context,
	t T,
	l zerolog.Logger,
	urlData downloadUrl,
	failedCh chan downloadUrl,
) {
	l.Trace().Int("idx", urlData.idx).Str("url", urlData.url).Msg("downloading page")

	err := c.infoProvider.DownloadContent(urlData.idx, t, urlData.url)
	if err == nil {
		time.Sleep(1 * time.Second)
		return
	}

	select {
	case <-innerCtx.Done():
		return
	case <-ctx.Done():
		return
	case failedCh <- urlData:
		c.failedDownloads++
		l.Warn().Err(err).Int("idx", urlData.idx).Str("url", urlData.url).
			Msg("download has failed for a page for the first time, trying page again at the end")
	}

	time.Sleep(1 * time.Second)
}

func (c *Core[T]) processFailedDownloads(
	innerCtx context.Context,
	ctx context.Context,
	t T,
	l zerolog.Logger,
	failedCh chan downloadUrl,
	errCh chan error,
	cancel context.CancelFunc,
) {
	for reTry := range failedCh {
		select {
		case <-innerCtx.Done():
			return
		case <-ctx.Done():
			return
		default:
			if err := c.infoProvider.DownloadContent(reTry.idx, t, reTry.url); err != nil {
				l.Error().Err(err).Str("url", reTry.url).Msg("Failed final download")
				select {
				case errCh <- fmt.Errorf("final download failed %w", err):
					cancel()
				default:
				}
				return
			}
			c.failedDownloads++
			time.Sleep(1 * time.Second)
		}
	}
}
