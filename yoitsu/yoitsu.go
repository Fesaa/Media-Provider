package yoitsu

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
	"log/slog"
	"math/rand"
	"net/http"
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
	client        *torrent.Client
	torrents      *utils.SafeMap[string, Torrent]
	clientBaseDir string
	baseDirs      *utils.SafeMap[string, string]
	dir           string
}

func (t *yoitsuImpl) GetBackingClient() *torrent.Client {
	return t.client
}

func (t *yoitsuImpl) GetBaseDir() string {
	return t.clientBaseDir
}

func (t *yoitsuImpl) AddDownload(infoHash string, baseDir string) (Torrent, error) {
	slog.Info("Adding torrent", "baseDir", baseDir, "infoHash", infoHash)
	torrentInfo, newTorrent := t.client.AddTorrentInfoHash(infohash.FromHexString(strings.ToLower(infoHash)))
	if !newTorrent {
		return nil, errors.New("torrent already exists")
	}

	return t.processTorrent(torrentInfo, baseDir), nil
}

func (t *yoitsuImpl) AddDownloadFromUrl(url string, baseDir string) (Torrent, error) {
	slog.Info("Adding torrent", "baseDir", baseDir, "url", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.New("failed to download torrent from url" + url + ": " + err.Error())
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.New("failed to get torrent file from url: " + url + " with status code: " + res.Status)
	}

	mi, err := metainfo.Load(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to load metainfo from url: %s, error: %s", url, err)
	}

	// client.AddTorrent starts downloading, so we need to add the baseDir to the map before calling it
	t.baseDirs.Set(mi.HashInfoBytes().HexString(), baseDir)

	torrentInfo, err := t.client.AddTorrent(mi)
	if err != nil {
		return nil, err
	}
	return t.processTorrent(torrentInfo, baseDir), nil
}

func (t *yoitsuImpl) processTorrent(torrentInfo *torrent.Torrent, dir string) Torrent {
	newTorrent := newTorrent(torrentInfo, dir)
	t.torrents.Set(torrentInfo.InfoHash().String(), newTorrent)
	t.baseDirs.Set(torrentInfo.InfoHash().String(), dir)
	newTorrent.WaitForInfoAndDownload()
	return newTorrent
}

func (t *yoitsuImpl) RemoveDownload(infoHashString string, deleteFiles bool) error {
	infoHashString = strings.ToLower(infoHashString)

	tor, ok := t.torrents.Get(infoHashString)
	if !ok {
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
		"deleteFiles", deleteFiles,
		"downloaded", backingTorrent.BytesCompleted(),
		"total", backingTorrent.Length())
	backingTorrent.Drop()

	t.torrents.Delete(infoHashString)
	t.baseDirs.Delete(infoHashString)
	if deleteFiles {
		go t.deleteTorrentFiles(backingTorrent, baseDir)
	} else {
		go t.cleanup(backingTorrent, baseDir)
	}
	return nil
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
		torrents:      utils.NewSafeMap[string, Torrent](),
		clientBaseDir: dir,
		baseDirs:      utils.NewSafeMap[string, string](),
		dir:           dir,
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
				err := t.RemoveDownload(s, false)
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
