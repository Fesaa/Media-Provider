package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"log/slog"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"
)

func NewDownloadableFromBlock[T any](req payload.DownloadRequest, block DownloadInfoProvider[T], client Client) *DownloadBase[T] {
	return &DownloadBase[T]{
		DownloadInfoProvider: block,
		Client:               client,
		Log:                  log.With(slog.String("id", req.Id)),
		id:                   req.Id,
		baseDir:              req.BaseDir,
		TempTitle:            req.TempTitle,
		maxImages:            min(client.GetConfig().GetMaxConcurrentImages(), 4),
		Req:                  req,
		LastTime:             time.Now(),
	}
}

type DownloadBase[T any] struct {
	DownloadInfoProvider[T]

	Client Client
	Log    *log.Logger

	id        string
	baseDir   string
	TempTitle string
	maxImages int
	Req       payload.DownloadRequest

	ToDownload      []T
	existingContent []string

	ContentDownloaded int
	ImagesDownloaded  int
	LastTime          time.Time
	LastRead          int

	ctx    context.Context
	cancel context.CancelFunc
	Wg     *sync.WaitGroup
}

func (d *DownloadBase[T]) Id() string {
	return d.id
}

func (d *DownloadBase[T]) GetBaseDir() string {
	return d.baseDir
}

func (d *DownloadBase[T]) GetDownloadDir() string {
	title := d.Title()
	if title == "" {
		return ""
	}
	return path.Join(d.baseDir, title)
}

func (d *DownloadBase[T]) GetOnDiskContent() []string {
	return d.existingContent
}

func (d *DownloadBase[T]) Cancel() {
	d.Log.Trace("calling cancel on manga")
	if d.cancel == nil {
		return
	}
	d.cancel()
	if d.Wg == nil {
		return
	}
	d.Wg.Wait()
}

func (d *DownloadBase[T]) WaitForInfoAndDownload() {
	if d.cancel != nil {
		d.Log.Debug("content already downloading")
		return
	}

	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.Log.Trace("loading content info")
	go func() {
		select {
		case <-d.ctx.Done():
			return
		case <-d.LoadInfo():
			d.Log = d.Log.With("title", d.Title())
			d.checkContentOnDisk()
			d.startDownload()
		}
	}()
}

func (d *DownloadBase[T]) checkContentOnDisk() {
	d.Log.Debug("checking content on disk", slog.String("dir", d.GetDownloadDir()))
	entries, err := os.ReadDir(path.Join(d.Client.GetBaseDir(), d.GetDownloadDir()))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d.Log.Trace("directory not found, fresh download")
		} else {
			d.Log.Warn("unable to check for already downloaded content. Downloading all", "err", err)
		}
		d.existingContent = []string{}
		return
	}

	out := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cbz") {
			d.Log.Trace("skipping non content file", "file", entry.Name())
			continue
		}

		matches := d.ContentRegex().FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}
		d.Log.Trace("found  content on disk",
			slog.String("file", entry.Name()),
			slog.String("key", matches[1]),
		)
		out = append(out, entry.Name())
	}

	d.Log.Debug("found following content on disk", "content", fmt.Sprintf("%v", out))
	d.existingContent = out
}

func (d *DownloadBase[T]) startDownload() {
	data := d.All()
	d.Log.Trace("starting download", slog.Int("size", len(data)))
	d.Wg = &sync.WaitGroup{}
	d.ToDownload = utils.Filter(data, func(t T) bool {
		download := !slices.Contains(d.existingContent, d.ContentDir(t)+".cbz")
		if !download {
			d.Log.Trace("content already downloaded, skipping", slog.String("key", d.ContentKey(t)))
		}
		return download
	})

	d.Log.Info("downloading content",
		slog.Int("all", len(data)),
		slog.Int("toDownload", len(d.ToDownload)))
	for _, content := range d.ToDownload {
		select {
		case <-d.ctx.Done():
			d.Wg.Wait()
			return
		default:
			d.Wg.Add(1)
			err := d.downloadContent(content)
			d.Wg.Done()
			if err != nil {
				d.Log.Error("error while downloading content; cleaning up", "err", err)
				req := payload.StopRequest{
					Provider:    d.Req.Provider,
					Id:          d.Id(),
					DeleteFiles: true,
				}
				if err = d.Client.RemoveDownload(req); err != nil {
					d.Log.Error("error while cleaning up", "err", err)
				}
				d.Wg.Wait()
				return
			}
		}
	}

	d.Wg.Wait()
	req := payload.StopRequest{
		Provider:    d.Req.Provider,
		Id:          d.Id(),
		DeleteFiles: false,
	}
	if err := d.Client.RemoveDownload(req); err != nil {
		d.Log.Error("error while cleaning up files", "err", err)
	}
}

func (d *DownloadBase[T]) downloadContent(t T) error {
	l := d.ContentLogger(t)

	l.Trace("downloading content")

	if err := os.MkdirAll(d.ContentPath(t), 0755); err != nil {
		return err
	}

	if err := d.WriteContentMetaData(t); err != nil {
		d.Log.Warn("error writing meta data", "err", err)
	}

	urls, err := d.ContentUrls(t)
	if err != nil {
		return err
	}
	l.Debug("downloading images", "size", len(urls))

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	sem := make(chan struct{}, d.maxImages)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, url := range urls {
		select {
		case <-d.ctx.Done():
			return nil
		case <-ctx.Done():
			wg.Wait()
			return errors.New("content download was cancelled from within")
		default:
			wg.Add(1)
			go func(i int, url string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				// Indexing pages from 1
				if err = d.DownloadContent(i+1, t, url); err != nil {
					select {
					case errCh <- err:
						cancel()
					default:
					}
				}
			}(i, url)
		}

		if (i+1)%d.maxImages == 0 && i > 0 {
			select {
			case <-time.After(1 * time.Second):
			case err := <-errCh:
				wg.Wait()
				for len(sem) > 0 {
					<-sem
				}
				return fmt.Errorf("encountered an error while downloading images: %w", err)
			case <-ctx.Done():
				wg.Wait()
				return fmt.Errorf("chapter download was cancelled from within")
			}
		}

		select {
		case err := <-errCh:
			wg.Wait()
			for len(sem) > 0 {
				<-sem
			}
			return fmt.Errorf("encountered an error while downloading images: %w", err)
		default:
		}
	}

	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}

	d.ContentDownloaded++
	return nil
}
