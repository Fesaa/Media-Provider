package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"
)

type IDAble interface {
	ID() string
}

func NewDownloadableFromBlock[T IDAble](req payload.DownloadRequest, block DownloadInfoProvider[T], client Client, log zerolog.Logger) *DownloadBase[T] {
	return &DownloadBase[T]{
		infoProvider: block,
		Client:       client,
		Log:          log.With().Str("id", req.Id).Logger(),
		id:           req.Id,
		baseDir:      req.BaseDir,
		TempTitle:    req.TempTitle,
		maxImages:    min(client.GetConfig().GetMaxConcurrentImages(), 4),
		Req:          req,
		LastTime:     time.Now(),
		ContentState: payload.ContentStateQueued,
	}
}

type Content struct {
	Name string
	Path string
}

type DownloadBase[T IDAble] struct {
	infoProvider DownloadInfoProvider[T]

	Client       Client
	Log          zerolog.Logger
	ContentState payload.ContentState

	id        string
	baseDir   string
	TempTitle string
	maxImages int
	Req       payload.DownloadRequest

	ToDownload      []T
	HasDownloaded   []string
	ExistingContent []Content

	// ToDownloadUserSelected are the ids of the content selected by the user to download in the UI
	ToDownloadUserSelected []string

	ContentDownloaded int
	ImagesDownloaded  int
	LastTime          time.Time
	LastRead          int

	ctx    context.Context
	cancel context.CancelFunc
	Wg     *sync.WaitGroup
}

func (d *DownloadBase[T]) Message(msg payload.Message) (payload.Message, error) {
	var jsonBytes []byte
	var err error
	switch msg.MessageType {
	case payload.MessageListContent:
		jsonBytes, err = json.Marshal(d.infoProvider.ContentList())
	case payload.SetToDownload:
		err = d.SetUserFiltered(msg.Data)
	case payload.StartDownload:
		err = d.MarkReady()
	default:
		return payload.Message{}, services.ErrUnknownMessageType
	}

	if err != nil {
		return payload.Message{}, err
	}

	return payload.Message{
		Provider:    d.Req.Provider,
		ContentId:   d.id,
		MessageType: msg.MessageType,
		Data:        jsonBytes,
	}, nil
}

func (d *DownloadBase[T]) MarkReady() error {
	if d.ContentState != payload.ContentStateWaiting {
		return services.ErrWrongState
	}

	if d.Client.CanStart(d.Req.Provider) {
		go d.StartDownload()
		return nil
	}

	d.ContentState = payload.ContentStateReady
	return nil
}

func (d *DownloadBase[T]) SetUserFiltered(msg json.RawMessage) error {
	if d.ContentState != payload.ContentStateWaiting &&
		d.ContentState != payload.ContentStateReady {
		return services.ErrWrongState
	}

	var filter []string
	err := json.Unmarshal(msg, &filter)
	if err != nil {
		return err
	}
	d.ToDownloadUserSelected = filter
	return nil
}

func (d *DownloadBase[T]) Id() string {
	return d.id
}

func (d *DownloadBase[T]) GetBaseDir() string {
	return d.baseDir
}

func (d *DownloadBase[T]) GetDownloadDir() string {
	title := d.infoProvider.Title()
	if title == "" {
		return ""
	}
	return path.Join(d.baseDir, title)
}

func (d *DownloadBase[T]) GetOnDiskContent() []Content {
	return d.ExistingContent
}

func (d *DownloadBase[T]) ExistingContentNames() []string {
	return utils.Map(d.ExistingContent, func(t Content) string {
		return t.Name
	})
}

func (d *DownloadBase[T]) GetContentByName(name string) (Content, bool) {
	for _, content := range d.ExistingContent {
		if content.Name == name {
			return content, true
		}
	}
	return Content{}, false
}

func (d *DownloadBase[T]) GetContentByPath(path string) (Content, bool) {
	for _, content := range d.ExistingContent {
		if content.Path == path {
			return content, true
		}
	}
	return Content{}, false
}

func (d *DownloadBase[T]) GetNewContent() []string {
	return d.HasDownloaded
}

func (d *DownloadBase[T]) Cancel() {
	d.Log.Trace().Msg("calling cancel on content")
	if d.cancel == nil {
		return
	}
	d.cancel()
	if d.Wg == nil {
		return
	}
	d.Wg.Wait()
}

func (d *DownloadBase[T]) StartLoadInfo() {
	if d.cancel != nil {
		d.Log.Debug().Msg("content already started")
		return
	}

	d.ContentState = payload.ContentStateLoading
	d.ctx, d.cancel = context.WithCancel(context.Background())
	d.Log.Debug().Msg("loading content info")
	select {
	case <-d.ctx.Done():
		return
	case <-d.infoProvider.LoadInfo():
	}

	d.Log = d.Log.With().Str("title", d.infoProvider.Title()).Logger()
	d.checkContentOnDisk()
	d.ContentState = utils.Ternary(d.Req.DownloadMetadata.StartImmediately,
		payload.ContentStateReady,
		payload.ContentStateWaiting)

	d.Log.Debug().Msg("Content has downloaded all information")

	data := d.infoProvider.All()
	d.ToDownload = utils.Filter(data, func(t T) bool {
		download := d.infoProvider.ShouldDownload(t)
		if !download {
			d.Log.Trace().Str("key", d.infoProvider.ContentKey(t)).Msg("content already downloaded, skipping")
		} else {
			d.Log.Trace().Str("key", d.infoProvider.ContentKey(t)).Msg("adding content to download queue")
		}
		return download
	})

	d.Log.Debug().Int("all", len(data)).Int("filtered", len(d.ToDownload)).
		Msg("downloaded content filtered")
}

func (d *DownloadBase[T]) StartDownload() {
	d.ContentState = payload.ContentStateDownloading
	go d.startDownload()
}

func (d *DownloadBase[T]) State() payload.ContentState {
	return d.ContentState
}

func (d *DownloadBase[T]) checkContentOnDisk() {
	d.Log.Debug().Str("dir", d.GetDownloadDir()).Msg("checking content on disk")
	content, err := d.readDirectoryForContent(d.GetDownloadDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			d.Log.Trace().Msg("directory not found, fresh download")
		} else {
			d.Log.Warn().Err(err).Msg("unable to check for already downloaded content. Downloading all")
		}
		d.ExistingContent = []Content{}
		return
	}

	d.Log.Debug().Str("content", fmt.Sprintf("%v", content)).Msg("found following content on disk")
	d.ExistingContent = content
}

func (d *DownloadBase[T]) readDirectoryForContent(p string) ([]Content, error) {
	entries, err := os.ReadDir(path.Join(d.Client.GetBaseDir(), p))
	if err != nil {
		return nil, err
	}

	out := make([]Content, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			dirContent, err2 := d.readDirectoryForContent(path.Join(p, entry.Name()))
			if err2 != nil {
				return nil, err2
			}
			out = append(out, dirContent...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".cbz") {
			d.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue

		}

		matches := d.infoProvider.IsContent(entry.Name())
		if !matches {
			d.Log.Trace().Str("file", entry.Name()).Msg("skipping non content file")
			continue
		}
		d.Log.Trace().Str("file", entry.Name()).Msg("found  content on disk")
		out = append(out, Content{
			Name: entry.Name(),
			Path: path.Join(p, entry.Name()),
		})
	}

	return out, nil
}

func (d *DownloadBase[T]) startDownload() {
	data := d.infoProvider.All()
	d.Log.Trace().Int("size", len(data)).Msg("downloading content")
	d.Wg = &sync.WaitGroup{}

	if len(d.ToDownloadUserSelected) > 0 {
		currentSize := len(d.ToDownload)
		d.ToDownload = utils.Filter(d.ToDownload, func(t T) bool {
			return slices.Contains(d.ToDownloadUserSelected, t.ID())
		})
		d.Log.Debug().Int("size", currentSize).Int("newSize", len(d.ToDownload)).
			Msg("content further filtered after user has made a selecting in the UI")
	}

	d.Log.Info().
		Int("all", len(data)).
		Int("toDownload", len(d.ToDownload)).
		Str("into", d.GetDownloadDir()).
		Msg("downloading content")
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
				d.Log.Error().Err(err).Msg("error while downloading content; cleaning up")
				req := payload.StopRequest{
					Provider:    d.Req.Provider,
					Id:          d.Id(),
					DeleteFiles: true,
				}
				if err = d.Client.RemoveDownload(req); err != nil {
					d.Log.Error().Err(err).Msg("error while cleaning up")
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
		d.Log.Error().Err(err).Msg("error while cleaning up files")
	}
}

//nolint:funlen,gocognit
func (d *DownloadBase[T]) downloadContent(t T) error {
	l := d.infoProvider.ContentLogger(t)

	l.Trace().Msg("downloading content")

	contentPath := d.infoProvider.ContentPath(t)
	if err := os.MkdirAll(contentPath, 0755); err != nil {
		return err
	}
	d.HasDownloaded = append(d.HasDownloaded, contentPath)

	urls, err := d.infoProvider.ContentUrls(t)
	if err != nil {
		return err
	}
	if len(urls) == 0 {
		l.Warn().Msg("content has no downloadable urls?")
		return nil
	}

	if err := d.infoProvider.WriteContentMetaData(t); err != nil {
		d.Log.Warn().Err(err).Msg("error writing meta data")
	}

	l.Debug().Int("size", len(urls)).Msg("downloading images")

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
				if err = d.infoProvider.DownloadContent(i+1, t, url); err != nil {
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
