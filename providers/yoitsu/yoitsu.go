package yoitsu

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"math/rand"
	"path"
	"strings"
	"time"
)

const (
	KeyNoSubDir string = "no-sub-dir"
)

type yoitsu struct {
	dir         string
	maxTorrents int

	client   *torrent.Client
	torrents utils.SafeMap[string, Torrent]
	baseDirs utils.SafeMap[string, string]

	log zerolog.Logger

	signalR    services.SignalRService
	notify     services.NotificationService
	dirService services.DirectoryService
	transLoco  services.TranslocoService
	fs         afero.Afero
}

func New(c *config.Config, log zerolog.Logger, signalR services.SignalRService,
	dirService services.DirectoryService, notify services.NotificationService,
	transLoco services.TranslocoService, fs afero.Afero, settingsService services.SettingsService,
) (Client, error) {
	settings, err := settingsService.GetSettingsDto()
	if err != nil {
		return nil, err
	}

	dir := utils.OrElse(settings.RootDir, "temp")

	impl := &yoitsu{
		dir:         dir,
		maxTorrents: settings.MaxConcurrentTorrents,

		torrents: utils.NewSafeMap[string, Torrent](),
		baseDirs: utils.NewSafeMap[string, string](),

		log:        log.With().Str("handler", "yoitsu").Logger(),
		signalR:    signalR,
		notify:     notify,
		dirService: dirService,
		transLoco:  transLoco,
		fs:         fs,
	}

	opts := storage.NewFileClientOpts{
		ClientBaseDir:   dir,
		TorrentDirMaker: impl.GetTorrentDirFilePathMaker(),
	}
	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewFileOpts(opts)
	conf.ListenPort = rand.Intn(65535-49152) + 49152 //nolint:gosec
	conf.DisableIPv6 = settings.DisableIpv6

	client, err := torrent.NewClient(conf)
	if err != nil {
		return nil, err
	}
	impl.client = client

	go impl.cleaner()
	return impl, nil
}

func (y *yoitsu) Shutdown() error {
	return nil
}

func (y *yoitsu) Content(id string) services.Content {
	content, ok := y.torrents.Get(id)
	if !ok {
		return nil
	}
	return content
}

func (y *yoitsu) GetBackingClient() *torrent.Client {
	return y.client
}

func (y *yoitsu) GetBaseDir() string {
	return y.dir
}

func (y *yoitsu) CanStartNext() bool {
	inUse := y.torrents.Count(func(k string, v Torrent) bool {
		return v.State() == payload.ContentStateDownloading || v.State() == payload.ContentStateLoading
	})

	return inUse < y.maxTorrents
}

func (y *yoitsu) Download(req payload.DownloadRequest) error {
	torrentInfo, nTorrent := y.client.AddTorrentInfoHash(infohash.FromHexString(strings.ToLower(req.Id)))
	if !nTorrent {
		return services.ErrContentAlreadyExists
	}

	torrentWrapper := newTorrent(torrentInfo, req, y.log, y, y.signalR, y.fs)
	y.torrents.Set(torrentInfo.InfoHash().String(), torrentWrapper)
	y.baseDirs.Set(torrentInfo.InfoHash().String(), req.BaseDir)
	y.signalR.AddContent(torrentWrapper.GetInfo())

	if !y.CanStartNext() {
		y.log.Debug().Msg("cannot start torrent, too many downloading")
		return nil
	}

	go func() {
		torrentWrapper.LoadInfo()

		if torrentWrapper.State() == payload.ContentStateReady {
			torrentWrapper.StartDownload()
		} else if torrentWrapper.State() == payload.ContentStateWaiting {
			y.log.Info().Str("infoHash", torrentWrapper.Id()).
				Str("title?", torrentWrapper.Title()).
				Msg("torrent is not ready for download, checking if an other can start")
			y.startNext()
		}
	}()

	return nil
}

func (y *yoitsu) RemoveDownload(req payload.StopRequest) error {
	infoHashString := strings.ToLower(req.Id)
	tor, ok := y.torrents.Get(infoHashString)
	if !ok {
		return services.ErrContentNotFound
	}

	// We may assume that it is present as long as the torrent is present
	baseDir, _ := y.baseDirs.Get(infoHashString)

	tor.Cancel()
	backingTorrent := tor.GetTorrent()
	y.log.Info().
		Str("name", backingTorrent.Name()).
		Str("infoHash", backingTorrent.InfoHash().HexString()).
		Bool("deleteFiles", req.DeleteFiles).
		Int64("downloaded", backingTorrent.BytesCompleted()).
		Int64("total", backingTorrent.Length()).
		Msg("dropping torrent")
	backingTorrent.Drop()

	y.torrents.Delete(infoHashString)
	y.baseDirs.Delete(infoHashString)

	y.signalR.StateUpdate(tor.Id(), payload.ContentStateCleanup)
	if req.DeleteFiles {
		go y.deleteTorrentFiles(tor, baseDir)
		go y.startNext()
		return nil
	}

	text := fmt.Sprintf("%s finished downloading %d files(s)", tor.Title(), tor.Files())
	y.notifier(tor.Request()).Notify(models.Notification{
		Title:   "Download finished",
		Summary: utils.Shorten(text, services.SummarySize),
		Body:    text,
		Colour:  models.Secondary,
		Group:   models.GroupContent,
	})

	go y.cleanup(tor, baseDir)
	go y.startNext()
	return nil
}

func (y *yoitsu) notifier(req payload.DownloadRequest) services.Notifier {
	if req.IsSubscription {
		return y.notify
	}

	return y.signalR
}

func (y *yoitsu) loadNext() {
	for y.CanStartNext() {
		inext, ok := y.torrents.Find(func(k string, v Torrent) bool {
			return v.State() == payload.ContentStateQueued
		})
		if !ok {
			return
		}

		next := *inext
		next.LoadInfo()
	}
}

func (y *yoitsu) startNext() {
	y.loadNext()

	inext, ok := y.torrents.Find(func(k string, v Torrent) bool {
		return v.State() == payload.ContentStateReady
	})

	if !ok {
		return
	}

	next := *inext
	next.StartDownload()

	if y.CanStartNext() {
		y.startNext()
	}
}

// TODO: Rewrite
//
//nolint:funlen
func (y *yoitsu) cleanup(t Torrent, baseDir string) {
	defer y.signalR.DeleteContent(t.Id())
	tor := t.GetTorrent()
	if tor == nil {
		return
	}

	var cleanupErrs []error
	defer func() {
		if len(cleanupErrs) > 0 {
			y.notifyCleanUpError(t, cleanupErrs...)
		}
	}()

	infoHash := tor.InfoHash().HexString()
	hashDir := path.Join(y.dir, baseDir, infoHash)
	info, err := y.fs.ReadDir(hashDir)
	if err != nil {
		y.log.Error().Err(err).Str("infoHash", infoHash).Str("dir", hashDir).Msg("error reading torrent dir")
		cleanupErrs = append(cleanupErrs, err)
		return
	}

	if len(info) == 0 {
		y.log.Warn().Msg("downloaded torrent was empty, removing directory")
		if err = y.fs.Remove(hashDir); err != nil {
			y.log.Error().Err(err).Str("dir", hashDir).Msg("error removing torrent dir")
			cleanupErrs = append(cleanupErrs, err)
		}
		return
	}

	// Calling cleanup beforehand, as it removed unwanted files. And the path might be incorrect after moving stuff
	t.Cleanup(hashDir)

	var src, dest string
	noSubDir := t.Request().GetBool(KeyNoSubDir, false)

	firstDirEntry := info[0]
	if len(info) == 1 && firstDirEntry.IsDir() {
		src = path.Join(hashDir, firstDirEntry.Name())
		dest = utils.Ternary(noSubDir, path.Join(y.dir, baseDir), path.Join(y.dir, baseDir, firstDirEntry.Name()))
		y.log.Debug().Str("infoHash", infoHash).Str("src", src).Str("dest", dest).
			Msg("torrent only has one directory, moving everything up")
	} else {
		src = hashDir
		dest = utils.Ternary(noSubDir, path.Join(y.dir, baseDir), path.Join(y.dir, baseDir, tor.Name()))
		y.log.Debug().Str("infoHash", infoHash).Str("src", src).Str("dest", dest).
			Msg("torrent downloaded more than one dirEntry, or a file; renaming directory")
	}

	if noSubDir {
		if err = y.dirService.MoveDirectoryContent(src, dest); err != nil {
			y.log.Error().Err(err).Str("src", src).Str("dest", dest).Msg("error moving directory contents")
			cleanupErrs = append(cleanupErrs, err)
			return
		}
	} else {
		if err = y.fs.Rename(src, dest); err != nil {
			y.log.Error().Err(err).Str("src", src).Str("dest", dest).Msg("error while renaming directory")
			cleanupErrs = append(cleanupErrs, err)
			return
		}
	}

	if len(info) == 1 && firstDirEntry.IsDir() {
		if err = y.fs.RemoveAll(hashDir); err != nil {
			y.log.Error().Err(err).Str("dir", hashDir).Msg("error removing torrent dir")
			cleanupErrs = append(cleanupErrs, err)
		}
	}
}

func (y *yoitsu) deleteTorrentFiles(tor Torrent, baseDir string) {
	if tor == nil {
		return
	}

	infoHash := tor.GetTorrent().InfoHash().HexString()
	defer y.signalR.DeleteContent(infoHash)

	dir := path.Join(y.dir, baseDir, infoHash)
	y.log.Debug().Str("infoHash", infoHash).Str("dir", dir).Msg("deleting directory")
	if err := y.fs.RemoveAll(dir); err != nil {
		y.log.Error().Err(err).Str("infoHash", infoHash).Str("dir", dir).Msg("error removing torrent dir")
		y.notifyCleanUpError(tor, err)
	}
}

func (y *yoitsu) GetTorrents() utils.SafeMap[string, Torrent] {
	return y.torrents
}

// GetTorrentDirFilePathMaker appending the infohash allows us to always clean up the torrent files on delete
func (y *yoitsu) GetTorrentDirFilePathMaker() storage.TorrentDirFilePathMaker {
	return func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
		d, ok := y.baseDirs.Get(infoHash.HexString())
		if !ok {
			return path.Join(baseDir, infoHash.HexString())
		}
		return path.Join(baseDir, d, infoHash.HexString())
	}
}

func (y *yoitsu) notifyCleanUpError(content Torrent, cleanupErrs ...error) {
	y.notify.NotifyContent(
		y.transLoco.GetTranslation("cleanup-errors-title"),
		y.transLoco.GetTranslation("cleanup-errors-summary", content.Title()),
		errors.Join(cleanupErrs...).Error(),
		models.Error)
}

func (y *yoitsu) cleaner() {
	for range time.Tick(time.Second * 5) {
		i := 0
		y.torrents.ForEach(func(s string, m Torrent) {
			if m.IsDone() {
				i++
				err := y.RemoveDownload(payload.StopRequest{
					Provider:    -1,
					Id:          s,
					DeleteFiles: false,
				})
				if err != nil {
					y.log.Error().Err(err).Str("file", s).Msg("error while cleaning up torrent")
				}
			}
		})
		if i > 0 {
			y.log.Trace().Int("amount", i).Msg("auto removing torrents")
		}
	}
}
