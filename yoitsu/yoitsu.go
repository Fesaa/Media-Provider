package yoitsu

import (
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

var yoitsu Yoitsu

func I() Yoitsu {
	return yoitsu
}

func Init() {
	var err error
	yoitsu, err = newYoitsu()
	if err != nil {
		slog.Error("Error initializing yoitsu:", "err", err)
		panic(err)
	}
}

type yoitsuImpl struct {
	client   *torrent.Client
	torrents *utils.SafeMap[string, Torrent]
	baseDirs *utils.SafeMap[string, string]
	dir      string
	queue    utils.Queue[payload.QueueStat]
}

func (t *yoitsuImpl) GetBackingClient() *torrent.Client {
	return t.client
}

func (t *yoitsuImpl) GetBaseDir() string {
	return t.dir
}

func (t *yoitsuImpl) AddDownload(req payload.DownloadRequest) (Torrent, error) {
	slog.Info("Adding torrent", "baseDir", req.BaseDir, "infoHash", req.Id)

	maxConcurrent := config.I().GetContentDownloaderConfig().GetMaxConcurrentTorrents()
	if maxConcurrent <= 0 {
		return t.addDownload(req)
	}
	if t.torrents.Len() >= maxConcurrent {
		t.queue.Enqueue(req.ToQueueStat())
		return nil, nil
	}

	return t.addDownload(req)
}

func (t *yoitsuImpl) addDownload(req payload.DownloadRequest) (Torrent, error) {
	torrentInfo, newTorrent := t.client.AddTorrentInfoHash(infohash.FromHexString(strings.ToLower(req.Id)))
	if !newTorrent {
		return nil, errors.New("torrent already exists")
	}
	return t.processTorrent(torrentInfo, req), nil
}

func (t *yoitsuImpl) processTorrent(torrentInfo *torrent.Torrent, req payload.DownloadRequest) Torrent {
	newTorrent := newTorrent(torrentInfo, req)
	t.torrents.Set(torrentInfo.InfoHash().String(), newTorrent)
	t.baseDirs.Set(torrentInfo.InfoHash().String(), req.BaseDir)
	newTorrent.WaitForInfoAndDownload()
	return newTorrent
}

func (t *yoitsuImpl) RemoveDownload(req payload.StopRequest) error {
	infoHashString := strings.ToLower(req.Id)
	tor, ok := t.torrents.Get(infoHashString)
	if !ok {
		ok = t.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == infoHashString
		})
		if ok {
			slog.Info("torrent removed from queue", "infoHash", infoHashString)
			return nil
		}

		return errors.New("torrent does not exist, or has already completed")
	}

	// We may assume that it is present as long as the torrent is present
	baseDir, _ := t.baseDirs.Get(infoHashString)

	err := tor.Cancel()
	if err != nil {
		slog.Error("Unable to cancel info loading", "err", err)
	}
	backingTorrent := tor.GetTorrent()
	slog.Info("Dropping torrent",
		"name", backingTorrent.Name(),
		"infoHash", backingTorrent.InfoHash().HexString(),
		"deleteFiles", req.DeleteFiles,
		"downloaded", backingTorrent.BytesCompleted(),
		"total", backingTorrent.Length())
	backingTorrent.Drop()

	t.torrents.Delete(infoHashString)
	t.baseDirs.Delete(infoHashString)
	if req.DeleteFiles {
		go t.deleteTorrentFiles(backingTorrent, baseDir)
	} else {
		go t.cleanup(backingTorrent, baseDir)
	}
	t.startNext()
	return nil
}

func (t *yoitsuImpl) startNext() {
	if t.queue.IsEmpty() {
		return
	}

	added := false
	for !added && !t.queue.IsEmpty() {
		item, _ := t.queue.Dequeue()
		_, err := t.AddDownload(item.ToDownloadRequest())
		if err != nil {
			slog.Warn("Error adding torrent from queue", "err", err)
			continue
		}
		added = true
	}
}

func (t *yoitsuImpl) cleanup(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}
	infoHash := tor.InfoHash().HexString()
	hashDir := path.Join(t.dir, baseDir, infoHash)
	info, err := os.ReadDir(hashDir)
	if err != nil {
		slog.Error("Error reading directory", "dir", hashDir, "err", err)
		return
	}

	if len(info) == 0 {
		slog.Warn("Downloaded torrent was empty, removing directory")
		if err := os.Remove(hashDir); err != nil {
			slog.Error("Error removing directory", "dir", hashDir, "err", err)
		}
		return
	}

	firstDirEntry := info[0]
	if len(info) == 1 && firstDirEntry.IsDir() {
		slog.Debug("Torrent only has one directory, moving everything up", "infoHash", infoHash)
		src := path.Join(hashDir, firstDirEntry.Name())
		dest := path.Join(t.dir, baseDir, firstDirEntry.Name())
		if err := os.Rename(src, dest); err != nil {
			slog.Error("Error renaming directory", "from", src, "to", dest, "err", err)
			return
		}
		if err := os.Remove(hashDir); err != nil {
			slog.Error("Error removing old hash directory", "dir", hashDir, "err", err)
		}
		return
	}

	slog.Debug("Torrent downloaded more than one dirEntry, or a file renaming directory", "infoHash", infoHash)
	src := hashDir
	dest := path.Join(t.dir, baseDir, tor.Name())
	if err := os.Rename(src, dest); err != nil {
		slog.Error("Error renaming directory", "from", src, "to", dest, "err", err)
		return
	}
}

func (t *yoitsuImpl) deleteTorrentFiles(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}

	infoHash := tor.InfoHash().HexString()
	dir := path.Join(t.dir, baseDir, infoHash)
	slog.Info("Deleting directory", "dir", dir, "infoHash", infoHash)
	err := os.RemoveAll(dir)
	if err != nil {
		slog.Error("Error deleting directory", "dir", dir, "err", err, "infoHash", infoHash)
	}
}

func (t *yoitsuImpl) GetRunningTorrents() *utils.SafeMap[string, Torrent] {
	return t.torrents
}

func (t *yoitsuImpl) GetQueuedTorrents() []payload.QueueStat {
	return t.queue.Items()
}

// GetTorrentDirFilePathMaker appending the infohash allows us to always clean up the torrent files on delete
func (t *yoitsuImpl) GetTorrentDirFilePathMaker() storage.TorrentDirFilePathMaker {
	return func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
		d, ok := t.baseDirs.Get(infoHash.HexString())
		if !ok {
			return path.Join(baseDir, infoHash.HexString())
		}
		return path.Join(baseDir, d, infoHash.HexString())
	}
}

func newYoitsu() (Yoitsu, error) {
	dir := config.OrDefault(config.I().GetRootDir(), "temp")

	impl := &yoitsuImpl{
		torrents: utils.NewSafeMap[string, Torrent](),
		baseDirs: utils.NewSafeMap[string, string](),
		dir:      dir,
		queue:    utils.NewQueue[payload.QueueStat](),
	}

	opts := storage.NewFileClientOpts{
		ClientBaseDir:   dir,
		TorrentDirMaker: impl.GetTorrentDirFilePathMaker(),
	}
	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewFileOpts(opts)
	conf.ListenPort = rand.Intn(65535-49152) + 49152

	client, err := torrent.NewClient(conf)
	if err != nil {
		return nil, err
	}
	impl.client = client

	go impl.cleaner()
	return impl, nil
}

func (t *yoitsuImpl) cleaner() {
	for range time.Tick(time.Second * 5) {
		i := 0
		t.torrents.ForEach(func(s string, m Torrent) {
			tor := m.GetTorrent()
			if tor.BytesCompleted() == tor.Length() && tor.BytesCompleted() > 0 {
				i++
				err := t.RemoveDownload(payload.StopRequest{
					Provider:    "",
					Id:          s,
					DeleteFiles: true,
				})
				if err != nil {
					slog.Error("Error removing torrent file", "file", s, "err", err)
				}
			}
		})
		if i > 0 {
			slog.Info("Removed torrent files", "files", i)
		}
	}
}
