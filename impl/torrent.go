package impl

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/Fesaa/Media-Provider/mount"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/types/infohash"
	"golang.org/x/sys/unix"
)

type TorrentImpl struct {
	client *torrent.Client

	torrents map[string]*models.Torrent

	clientBaseDir string
	baseDirs      map[string]string
	lock          *sync.RWMutex
	lockDir       *sync.RWMutex
	dir           string
}

func (t *TorrentImpl) GetBackingClient() *torrent.Client {
	return t.client
}

func (t *TorrentImpl) AddDownload(infoHash string, baseDir string) (*models.Torrent, error) {
	err := t.checkMount(baseDir)
	if err != nil {
		return nil, err
	}

	torrentInfo, new := t.client.AddTorrentInfoHash(infohash.FromHexString(strings.ToLower(infoHash)))
	if !new {
		return nil, errors.New("torrent already exists")
	}

	return t.processTorrent(torrentInfo, baseDir), nil
}

func (t *TorrentImpl) AddDownloadFromUrl(url string, baseDir string) (*models.Torrent, error) {
	err := t.checkMount(baseDir)
	if err != nil {
		return nil, err
	}

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		slog.Error("Failed to get torrent file from url: " + url + " with status code: " + res.Status)
		return nil, errors.New("failed to get torrent file from url: " + url + " with status code: " + res.Status)
	}

	mi, err := metainfo.Load(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to load metainfo from url: %s, error: %s", url, err)
	}

	// client.AddTorrent starts downloading, so we need to add the baseDir to the map before calling it
	safeSet(t.baseDirs, mi.HashInfoBytes().HexString(), baseDir, t.lockDir)

	torrentInfo, err := t.client.AddTorrent(mi)
	if err != nil {
		return nil, err
	}
	return t.processTorrent(torrentInfo, baseDir), nil
}

func (t *TorrentImpl) processTorrent(torrentInfo *torrent.Torrent, dir string) *models.Torrent {
	torrent := models.NewTorrent(torrentInfo, dir)
	safeSet(t.torrents, torrentInfo.InfoHash().String(), torrent, t.lock)
	safeSet(t.baseDirs, torrentInfo.InfoHash().String(), dir, t.lockDir)

	go func() {
		<-torrentInfo.GotInfo()
		slog.Info(fmt.Sprintf("Info received for %s, starting download.", torrentInfo.Name()))
		torrentInfo.DownloadAll()
	}()
	return torrent
}

func (t *TorrentImpl) checkMount(baseDir string) error {
	if !mount.WantsMount() {
		return nil
	}

	if len(t.torrents) == 0 && !writable(t.clientBaseDir+"/"+baseDir) {
		slog.Info("Base dir is not writable, attempting remount.")
		mount.Unmount()
		mount.Mount(false)
		slog.Info("Remounted, checking if base dir is writable.")
		if !writable(t.clientBaseDir + "/" + baseDir) {
			go func() {
				time.Sleep(time.Duration(10) * time.Second)
				os.Exit(1)
			}()
			return errors.New("base dir is not writable after remount, exiting program soon")
		}
	}

	return nil
}

func (t *TorrentImpl) RemoveDownload(infoHashString string, deleteFiles bool) error {
	infoHashString = strings.ToLower(infoHashString)

	tor, ok := safeGet(t.torrents, infoHashString, t.lock)
	if !ok {
		return errors.New("torrent does not exist, or has already completed")
	}

	// We may assume that it is present as long as the torrent is present
	baseDir, _ := safeGet(t.baseDirs, infoHashString, t.lockDir)

	torrent := tor.GetTorrent()
	slog.Info(fmt.Sprintf("Dropping torrent %s. Had %d / %d", torrent.Name(), torrent.BytesCompleted(), torrent.Length()))
	torrent.Drop()

	safeDelete(t.torrents, infoHashString, t.lock)
	safeDelete(t.baseDirs, infoHashString, t.lockDir)
	if deleteFiles {
		t.deleteTorrentFiles(torrent, baseDir)
	}
	return nil
}

func (t *TorrentImpl) deleteTorrentFiles(tor *torrent.Torrent, baseDir string) {
	if tor == nil {
		return
	}

	dir := t.dir + "/" + baseDir + "/" + tor.InfoHash().HexString()
	slog.Info(fmt.Sprintf("Deleting dir %s", dir))
	err := os.RemoveAll(dir)
	if err != nil {
		slog.Error(fmt.Sprintf("Error deleting dir %s: %s", dir, err))
	}
}

func (t *TorrentImpl) GetRunningTorrents() map[string]*models.Torrent {
	return t.torrents
}

// Appending the infohash allows us to always cleanup the torrent files on delete
// This does however mean that if the torrent has it's own upper dir, it'll be layered
func (t *TorrentImpl) GetTorrentDirFilePathMaker() storage.TorrentDirFilePathMaker {
	return func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
		d, ok := safeGet(t.baseDirs, infoHash.HexString(), t.lockDir)
		if !ok {
			return baseDir + "/" + infoHash.HexString()
		}
		return baseDir + "/" + d + "/" + infoHash.HexString()
	}
}

func newTorrent() (*TorrentImpl, error) {
	dir := utils.GetEnv("TORRENT_DIR", "temp")

	impl := &TorrentImpl{
		torrents: make(map[string]*models.Torrent),

		clientBaseDir: dir,
		baseDirs:      make(map[string]string),
		lock:          &sync.RWMutex{},
		lockDir:       &sync.RWMutex{},
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

func (t *TorrentImpl) cleaner() {
	for range time.Tick(time.Second * 5) {
		t.lock.RLock()
		i := 0
		for infoHash, tor := range t.torrents {
			torrent := tor.GetTorrent()
			if torrent.BytesCompleted() == torrent.Length() && torrent.BytesCompleted() > 0 {
				i++
				t.lock.RUnlock()
				t.RemoveDownload(infoHash, false)
				t.lock.RLock()
			}
		}
		if i > 0 {
			slog.Info(fmt.Sprintf("Removed %d completed torrents", i))
		}
		t.lock.RUnlock()
	}
}

func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func safeGet[K comparable, T any](m map[K]T, k K, lock *sync.RWMutex) (T, bool) {
	lock.RLock()
	defer lock.RUnlock()
	v, ok := m[k]
	return v, ok
}

func safeSet[K comparable, T any](m map[K]T, k K, v T, lock *sync.RWMutex) {
	lock.Lock()
	defer lock.Unlock()
	m[k] = v
}

func safeDelete[K comparable, T any](m map[K]T, k K, lock *sync.RWMutex) {
	lock.Lock()
	defer lock.Unlock()
	delete(m, k)
}
