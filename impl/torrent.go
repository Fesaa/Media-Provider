package impl

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/types/infohash"
)

type TorrentImpl struct {
	client *torrent.Client

	torrents map[string]*torrent.Torrent
	lock     *sync.RWMutex
}

func (t *TorrentImpl) GetBackingClient() *torrent.Client {
	return t.client
}

func (t *TorrentImpl) AddDownload(infoHashString string) (*torrent.Torrent, error) {
	infoHash := infohash.FromHexString(infoHashString)
	torrentInfo, new := t.client.AddTorrentInfoHash(infoHash)
	if !new {
		return nil, errors.New("torrent already exists")
	}

	t.lock.Lock()
	t.torrents[infoHashString] = torrentInfo
	t.lock.Unlock()

	go func() {
		<-torrentInfo.GotInfo()
		slog.Info(fmt.Sprintf("Info received for %s, starting download.", torrentInfo.Name()))
		torrentInfo.DownloadAll()
	}()

	return torrentInfo, nil
}

func (t *TorrentImpl) RemoveDownload(infoHashString string) error {
	t.lock.RLock()
	torrent, ok := t.torrents[infoHashString]
	t.lock.RUnlock()
	if !ok {
		return errors.New("torrent does not exist, or has already completed")
	}

	slog.Info(fmt.Sprintf("Dropping torrent %s. Had %d / %d", torrent.Name(), torrent.BytesCompleted(), torrent.Length()))
	torrent.Drop()
	t.lock.Lock()
	delete(t.torrents, infoHashString)
	t.lock.Unlock()
	return nil
}

func (t *TorrentImpl) GetRunningTorrents() map[string]*torrent.Torrent {
	return t.torrents
}

func newTorrent(c *torrent.ClientConfig) (*TorrentImpl, error) {
	client, err := torrent.NewClient(c)
	if err != nil {
		return nil, err
	}

	impl := &TorrentImpl{
		client:   client,
		torrents: make(map[string]*torrent.Torrent),
		lock:     &sync.RWMutex{},
	}
	go impl.cleaner()

	return impl, nil
}

func (t *TorrentImpl) cleaner() {
	for range time.Tick(time.Second * 5) {
		t.lock.RLock()
		i := 0
		for infoHash, torrent := range t.torrents {
			if torrent.BytesCompleted() == torrent.Length() {
				i++
				t.lock.RUnlock()
				t.RemoveDownload(infoHash)
				t.lock.RLock()
			}
		}
		if i > 0 {
			slog.Info(fmt.Sprintf("Removed %d completed torrents", i))
		}
		t.lock.RUnlock()
	}
}
