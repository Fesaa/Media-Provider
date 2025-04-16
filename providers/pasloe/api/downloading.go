package api

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

func (d *DownloadBase[T]) abortDownload(reason error) {
	if errors.Is(reason, context.Canceled) {
		return
	}

	d.Log.Error().Err(reason).Msg("error while downloading content; cleaning up")
	req := payload.StopRequest{
		Provider:    d.Req.Provider,
		Id:          d.Id(),
		DeleteFiles: true,
	}
	if err := d.Client.RemoveDownload(req); err != nil {
		d.Log.Error().Err(err).Msg("error while cleaning up")
	}
	d.Notifier.Notify(models.Notification{
		Title:   "Failed download",
		Summary: fmt.Sprintf("%s failed to download", d.infoProvider.Title()),
		Body:    fmt.Sprintf("Download failed for %s, because %v", d.infoProvider.Title(), reason),
		Colour:  models.Red,
		Group:   models.GroupError,
	})
}

func (d *DownloadBase[T]) filterContentByUserSelection() {
	if len(d.ToDownloadUserSelected) > 0 {
		currentSize := len(d.ToDownload)
		d.ToDownload = utils.Filter(d.ToDownload, func(t T) bool {
			return slices.Contains(d.ToDownloadUserSelected, t.ID())
		})
		d.Log.Debug().Int("size", currentSize).Int("newSize", len(d.ToDownload)).
			Msg("content further filtered after user has made a selection in the UI")

		if len(d.ToRemoveContent) > 0 {
			paths := utils.Map(d.ToDownload, func(t T) string {
				return d.infoProvider.ContentPath(t) + ".cbz"
			})
			d.ToRemoveContent = utils.Filter(d.ToRemoveContent, func(s string) bool {
				return slices.Contains(paths, s)
			})
		}
	}
}

func (d *DownloadBase[T]) processDownloads(ctx context.Context) error {
	for _, content := range d.ToDownload {
		select {
		case <-ctx.Done():
			d.Wg.Wait()
			return errors.New("download cancelled")
		default:
			d.Wg.Add(1)
			err := d.downloadContent(ctx, content)
			d.Wg.Done()
			if err != nil {
				d.abortDownload(err)
				d.Wg.Wait()
				return err
			}
		}
		d.UpdateProgress()
	}
	return nil
}

func (d *DownloadBase[T]) cleanupAfterDownload() {
	d.Wg.Wait()
	req := payload.StopRequest{
		Provider:    d.Req.Provider,
		Id:          d.Id(),
		DeleteFiles: false,
	}
	if err := d.Client.RemoveDownload(req); err != nil {
		d.Log.Error().Err(err).Msg("error while cleaning up files")
	}
}

func (d *DownloadBase[T]) startDownload() {
	// Overwrite cancel, as we're doing something else
	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel

	data := d.infoProvider.All()
	d.Log.Trace().Int("size", len(data)).Msg("downloading content")
	d.Wg = &sync.WaitGroup{}

	d.filterContentByUserSelection()

	d.Log.Info().
		Int("all", len(data)).
		Int("toDownload", len(d.ToDownload)).
		Int("reDownloads", len(d.ToRemoveContent)).
		Str("into", d.GetDownloadDir()).
		Msg("downloading content")

	if err := d.processDownloads(ctx); err != nil {
		d.Log.Trace().Err(err).Msg("download failed")
		return
	}

	d.cleanupAfterDownload()
}

type downloadUrl struct {
	idx int
	url string
}

func (d *DownloadBase[T]) downloadContent(ctx context.Context, t T) error {
	l := d.infoProvider.ContentLogger(t)
	l.Trace().Msg("downloading content")

	contentPath := d.infoProvider.ContentPath(t)
	if err := d.fs.MkdirAll(contentPath, 0755); err != nil {
		return err
	}
	d.HasDownloaded = append(d.HasDownloaded, contentPath)

	urls, err := d.infoProvider.ContentUrls(ctx, t)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		l.Warn().Msg("content has no downloadable urls? Unexpected? Report this!")
		return nil
	}

	if err = d.infoProvider.WriteContentMetaData(t); err != nil {
		d.Log.Warn().Err(err).Msg("error writing meta data")
	}

	l.Debug().Int("size", len(urls)).Msg("downloading images")

	urlCh := make(chan downloadUrl, d.maxImages)
	errCh := make(chan error, 1)
	defer close(errCh)

	wg := &sync.WaitGroup{}
	innerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d.produceURLs(ctx, innerCtx, urls, urlCh)

	d.startProgressUpdater(ctx, innerCtx)

	for range d.maxImages {
		wg.Add(1)
		go d.channelConsumer(innerCtx, cancel, ctx, t, l, urlCh, errCh, wg)
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

	d.ContentDownloaded++
	return nil
}

func (d *DownloadBase[T]) produceURLs(ctx context.Context, innerCtx context.Context, urls []string, urlCh chan<- downloadUrl) {
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

func (d *DownloadBase[T]) startProgressUpdater(ctx context.Context, innerCtx context.Context) {
	go func() {
		for range time.Tick(2 * time.Second) {
			select {
			case <-innerCtx.Done():
				return
			case <-ctx.Done():
				return
			default:
				d.UpdateProgress()
			}
		}
	}()
}

func (d *DownloadBase[T]) channelConsumer(
	innerCtx context.Context,
	cancel context.CancelFunc,
	ctx context.Context,
	t T,
	l zerolog.Logger,
	urlCh chan downloadUrl,
	errCh chan error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	failedCh := make(chan downloadUrl)

	d.processInitialDownloads(innerCtx, ctx, t, l, urlCh, failedCh)
	d.processFailedDownloads(innerCtx, ctx, t, l, failedCh, errCh, cancel)
}

func (d *DownloadBase[T]) processInitialDownloads(
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
			d.downloadURL(innerCtx, ctx, t, l, urlData, failedCh)
		}
	}
	close(failedCh)
}

func (d *DownloadBase[T]) downloadURL(
	innerCtx context.Context,
	ctx context.Context,
	t T,
	l zerolog.Logger,
	urlData downloadUrl,
	failedCh chan downloadUrl,
) {
	l.Trace().Int("idx", urlData.idx).Str("url", urlData.url).Msg("downloading page")

	err := d.infoProvider.DownloadContent(urlData.idx, t, urlData.url)
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
		d.failedDownloads++
		l.Warn().Err(err).Int("idx", urlData.idx).Str("url", urlData.url).
			Msg("download has failed for a page for the first time, trying page again at the end")
	}

	time.Sleep(1 * time.Second)
}

func (d *DownloadBase[T]) processFailedDownloads(
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
			if err := d.infoProvider.DownloadContent(reTry.idx, t, reTry.url); err != nil {
				l.Error().Err(err).Str("url", reTry.url).Msg("Failed final download")
				select {
				case errCh <- fmt.Errorf("final download failed %w", err):
					cancel()
				default:
				}
				return
			}
			d.failedDownloads++
			time.Sleep(1 * time.Second)
		}
	}
}
