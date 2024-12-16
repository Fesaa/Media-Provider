package yoitsu

import (
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
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

func Init(c Config) {
	var err error
	yoitsu, err = newYoitsu(c)
	if err != nil {
		log.Fatal("error while initializing Yoitsu", err)
		panic(err)
	}
}

type yoitsuImpl struct {
	dir         string
	maxTorrents int

	client   *torrent.Client
	torrents *utils.SafeMap[string, Torrent]
	baseDirs *utils.SafeMap[string, string]
	queue    utils.Queue[payload.QueueStat]
}

func newYoitsu(c Config) (Yoitsu, error) {
	dir := config.OrDefault(c.GetRootDir(), "temp")

	impl := &yoitsuImpl{
		dir:         dir,
		maxTorrents: c.GetMaxConcurrentTorrents(),

		torrents: utils.NewSafeMap[string, Torrent](),
		baseDirs: utils.NewSafeMap[string, string](),
		queue:    utils.NewQueue[payload.QueueStat](),
	}

	opts := storage.NewFileClientOpts{
		ClientBaseDir:   dir,
		TorrentDirMaker: impl.GetTorrentDirFilePathMaker(),
	}
	conf := torrent.NewDefaultClientConfig()
	conf.DefaultStorage = storage.NewFileOpts(opts)
	conf.ListenPort = rand.Intn(65535-49152) + 49152
	conf.DisableIPv6 = config.I().Downloader.DisableIpv6

	client, err := torrent.NewClient(conf)
	if err != nil {
		return nil, err
	}
	impl.client = client

	go impl.cleaner()
	return impl, nil
}

func (y *yoitsuImpl) GetBackingClient() *torrent.Client {
	return y.client
}

func (y *yoitsuImpl) GetBaseDir() string {
	return y.dir
}

func (y *yoitsuImpl) AddDownload(req payload.DownloadRequest) (Torrent, error) {
	if y.maxTorrents <= 0 {
		return y.addDownload(req)
	}
	if y.torrents.Len() >= y.maxTorrents {
		y.queue.Enqueue(req.ToQueueStat())
		return nil, nil
	}

	log.Info("downloading torrent", "infoHash", req.Id, "into", req.BaseDir, "title?", req.TempTitle)
	return y.addDownload(req)
}

func (y *yoitsuImpl) addDownload(req payload.DownloadRequest) (Torrent, error) {
	torrentInfo, nTorrent := y.client.AddTorrentInfoHash(infohash.FromHexString(strings.ToLower(req.Id)))
	if !nTorrent {
		return nil, errors.New("torrent already exists")
	}
	return y.processTorrent(torrentInfo, req), nil
}

func (y *yoitsuImpl) processTorrent(torrentInfo *torrent.Torrent, req payload.DownloadRequest) Torrent {
	nTorrent := newTorrent(torrentInfo, req)
	y.torrents.Set(torrentInfo.InfoHash().String(), nTorrent)
	y.baseDirs.Set(torrentInfo.InfoHash().String(), req.BaseDir)
	nTorrent.WaitForInfoAndDownload()
	return nTorrent
}

func (y *yoitsuImpl) RemoveDownload(req payload.StopRequest) error {
	infoHashString := strings.ToLower(req.Id)
	tor, ok := y.torrents.Get(infoHashString)
	if !ok {
		ok = y.queue.RemoveFunc(func(item payload.QueueStat) bool {
			return item.Id == infoHashString
		})
		if ok {
			log.Info("torrent removed from queue", "infoHash", infoHashString)
			return nil
		}

		return errors.New("torrent does not exist, or has already completed")
	}

	// We may assume that it is present as long as the torrent is present
	baseDir, _ := y.baseDirs.Get(infoHashString)

	err := tor.Cancel()
	if err != nil {
		log.Error("error while canceling torrent", "err", err)
	}
	backingTorrent := tor.GetTorrent()
	log.Info("dropping torrent",
		"name", backingTorrent.Name(),
		"infoHash", backingTorrent.InfoHash().HexString(),
		"deleteFiles", req.DeleteFiles,
		"downloaded", backingTorrent.BytesCompleted(),
		"total", backingTorrent.Length())
	backingTorrent.Drop()

	y.torrents.Delete(infoHashString)
	y.baseDirs.Delete(infoHashString)
	if req.DeleteFiles {
		go y.deleteTorrentFiles(backingTorrent, baseDir)
	} else {
		go y.cleanup(backingTorrent, baseDir)
	}
	y.startNext()
	return nil
}

func (y *yoitsuImpl) startNext() {
	if y.queue.IsEmpty() {
		return
	}

	added := false
	for !added && !y.queue.IsEmpty() {
		item, _ := y.queue.Dequeue()
		_, err := y.AddDownload(item.ToDownloadRequest())
		if err != nil {
			log.Warn("error while adding torrent from queue", "err", err)
			continue
		}
		added = true
	}
}

func (y *yoitsuImpl) cleanup(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}
	infoHash := tor.InfoHash().HexString()
	hashDir := path.Join(y.dir, baseDir, infoHash)
	info, err := os.ReadDir(hashDir)
	if err != nil {
		log.Error("error while reading directory", "dir", hashDir, "infoHash", infoHash, "err", err)
		return
	}

	if len(info) == 0 {
		log.Warn("downloaded torrent was empty, removing directory")
		if err = os.Remove(hashDir); err != nil {
			log.Error("error while removing directory", "dir", hashDir, "err", err)
		}
		return
	}

	firstDirEntry := info[0]
	if len(info) == 1 && firstDirEntry.IsDir() {
		log.Debug("torrent only has one directory, moving everything up", "infoHash", infoHash)
		src := path.Join(hashDir, firstDirEntry.Name())
		dest := path.Join(y.dir, baseDir, firstDirEntry.Name())
		if err = os.Rename(src, dest); err != nil {
			log.Error("error while renaming directory", "from", src, "to", dest, "err", err)
			return
		}

		if err = os.RemoveAll(hashDir); err != nil {
			log.Error("error while deleting directory", "dir", hashDir, "err", err, "infoHash", infoHash)
		}
		return
	}

	log.Debug("torrent downloaded more than one dirEntry, or a file; renaming directory", "infoHash", infoHash)
	src := hashDir
	dest := path.Join(y.dir, baseDir, tor.Name())
	if err = os.Rename(src, dest); err != nil {
		log.Error("error while renaming directory", "from", src, "to", dest, "err", err)
		return
	}
}

func (y *yoitsuImpl) deleteTorrentFiles(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}

	infoHash := tor.InfoHash().HexString()
	dir := path.Join(y.dir, baseDir, infoHash)
	log.Debug("deleting directory", "dir", dir, "infoHash", infoHash)
	if err := os.RemoveAll(dir); err != nil {
		log.Error("error while deleting directory", "dir", dir, "err", err, "infoHash", infoHash)
	}
}

func (y *yoitsuImpl) GetRunningTorrents() *utils.SafeMap[string, Torrent] {
	return y.torrents
}

func (y *yoitsuImpl) GetQueuedTorrents() []payload.QueueStat {
	return y.queue.Items()
}

// GetTorrentDirFilePathMaker appending the infohash allows us to always clean up the torrent files on delete
func (y *yoitsuImpl) GetTorrentDirFilePathMaker() storage.TorrentDirFilePathMaker {
	return func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
		d, ok := y.baseDirs.Get(infoHash.HexString())
		if !ok {
			return path.Join(baseDir, infoHash.HexString())
		}
		return path.Join(baseDir, d, infoHash.HexString())
	}
}

func (y *yoitsuImpl) cleaner() {
	for range time.Tick(time.Second * 5) {
		i := 0
		y.torrents.ForEach(func(s string, m Torrent) {
			tor := m.GetTorrent()
			if tor.BytesCompleted() == tor.Length() && tor.BytesCompleted() > 0 {
				i++
				err := y.RemoveDownload(payload.StopRequest{
					Provider:    -1,
					Id:          s,
					DeleteFiles: false,
				})
				if err != nil {
					log.Error("error while cleaning up torrent", "file", s, "err", err)
				}
			}
		})
		if i > 0 {
			log.Trace("auto removing torrents", "amount", i)
		}
	}
}
