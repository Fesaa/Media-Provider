package impl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

type TorrentImpl struct {
	client *torrent.Client

	torrents map[string]*models.Torrent
	lock     *sync.RWMutex
	dir      string
}

func (t *TorrentImpl) GetBackingClient() *torrent.Client {
	return t.client
}

func (t *TorrentImpl) AddDownload(infoHashString string) (*models.Torrent, error) {
	infoHash := infohash.FromHexString(infoHashString)
	torrentInfo, new := t.client.AddTorrentInfoHash(infoHash)
	if !new {
		return nil, errors.New("torrent already exists")
	}

	torrent := models.NewTorrent(torrentInfo)
	t.lock.Lock()
	t.torrents[infoHashString] = torrent
	t.lock.Unlock()

	go func() {
		<-torrentInfo.GotInfo()
		slog.Info(fmt.Sprintf("Info received for %s, starting download.", torrentInfo.Name()))
		torrentInfo.DownloadAll()
	}()

	return torrent, nil
}

func (t *TorrentImpl) RemoveDownload(infoHashString string, deleteFiles bool) error {
	t.lock.RLock()
	tor, ok := t.torrents[infoHashString]
	t.lock.RUnlock()
	if !ok {
		return errors.New("torrent does not exist, or has already completed")
	}

	torrent := tor.GetTorrent()
	slog.Info(fmt.Sprintf("Dropping torrent %s. Had %d / %d", torrent.Name(), torrent.BytesCompleted(), torrent.Length()))
	torrent.Drop()
	t.lock.Lock()
	delete(t.torrents, infoHashString)
	t.lock.Unlock()
	if deleteFiles {
		t.deleteTorrentFiles(torrent)
	}
	return nil
}

// This doesn't delete the dir the files are downloaded at. Only the files inside it.
// Not sure how to get the correct dir all the time, so leaving it be for now.
func (t *TorrentImpl) deleteTorrentFiles(tor *torrent.Torrent) {
	if tor == nil {
		return
	}
	for _, file := range tor.Files() {
		// I/O should be done async to speed it up
		// Passing the file, so it's not scope locked
		go func(file *torrent.File) {
			path := t.dir + "/" + file.Path()
			err := os.Remove(path)
			if err != nil {
				// This will log quite a bit if the torrent is deleted early on.
				// This is fine, we still want to log all of it for visibility in case a file is not deleted.
				slog.Error(fmt.Sprintf("Error deleting file %s: %s", path, err))
			}
		}(file)
	}
}

func (t *TorrentImpl) GetRunningTorrents() map[string]*models.Torrent {
	return t.torrents
}

func newTorrent(c *torrent.ClientConfig) (*TorrentImpl, error) {
	client, err := torrent.NewClient(c)
	if err != nil {
		return nil, err
	}

	impl := &TorrentImpl{
		client:   client,
		torrents: make(map[string]*models.Torrent),
		lock:     &sync.RWMutex{},
		dir:      c.DataDir,
	}
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
