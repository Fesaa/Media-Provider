package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
)

type SpeedData struct {
	t     time.Time
	bytes int64
}

type Torrent struct {
	t       *torrent.Torrent
	key     string
	baseDir string

	lock     *sync.RWMutex
	speedMap map[string]SpeedData
}

type TorrentInfo struct {
	InfoHash  string `json:"InfoHash"`
	Name      string `json:"Name"`
	Size      int64  `json:"Size"`
	Progress  int64  `json:"Progress"`
	Completed int64  `json:"Completed"`
	Speed     string `json:"Speed"`
}

type TorrentProvider interface {
	GetBackingClient() *torrent.Client

	AddDownload(infoHash string, baseDir string) (*Torrent, error)
	RemoveDownload(infoHash string, deleteFiles bool) error
	GetRunningTorrents() map[string]*Torrent
}

func NewTorrent(t *torrent.Torrent, baseDir string) *Torrent {
	return &Torrent{
		t:        t,
		key:      t.InfoHash().HexString(),
		baseDir:  baseDir,
		lock:     &sync.RWMutex{},
		speedMap: make(map[string]SpeedData),
	}
}

func (t *Torrent) GetTorrent() *torrent.Torrent {
	return t.t
}

func (t *Torrent) GetInfo() TorrentInfo {
	c := t.t.Stats().BytesReadData
	progress := c.Int64()
	var speed int64 = 0

	t.lock.RLock()
	s, ok := t.speedMap[t.key]
	t.lock.RUnlock()
	if ok {
		bytesDiff := progress - s.bytes
		timeDiff := time.Since(s.t).Seconds()
		speed = int64(float64(bytesDiff) / timeDiff)
	}

	t.lock.Lock()
	t.speedMap[t.key] = SpeedData{
		t:     time.Now(),
		bytes: progress,
	}
	t.lock.Unlock()

	return TorrentInfo{
		InfoHash:  t.key,
		Name:      t.t.Name(),
		Size:      t.t.Length(),
		Progress:  progress,
		Completed: percent(progress, t.t.Length()),
		Speed:     humanReadableSpeed(speed),
	}
}

func humanReadableSpeed(speed int64) string {
	if speed < 1024 {
		return fmt.Sprintf("%d B/s", speed)
	}
	speed /= 1024
	if speed < 1024 {
		return fmt.Sprintf("%d KB/s", speed)
	}
	speed /= 1024
	return fmt.Sprintf("%d MB/s", speed)
}

func percent(a, b int64) int64 {
	b = max(b, 1)
	ratio := (float64)(a) / (float64)(b)
	return (int64)(ratio * 100)
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
