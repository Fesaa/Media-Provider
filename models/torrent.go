package models

import (
	"fmt"
	"time"

	"github.com/anacrolix/torrent"
)

// Contains the amount of download bytes at a certain time
type SpeedData struct {
	t     time.Time
	bytes int64
}

// Wrapper around the torrent.Torrent struct
// Providers some specific functionality
type Torrent struct {
	t       *torrent.Torrent
	key     string
	baseDir string

	lastSpeed SpeedData
}

// Contains information about a torrent
// This is the information that is sent to the client
type TorrentInfo struct {
	InfoHash  string `json:"InfoHash"`
	Name      string `json:"Name"`
	Size      int64  `json:"Size"`
	Progress  int64  `json:"Progress"`
	Completed int64  `json:"Completed"`
	Speed     string `json:"Speed"`
}

// Wrapper around the torrent.Client struct
type TorrentProvider interface {
	// Returns the the torrent.Client
	GetBackingClient() *torrent.Client

	// Adds a new download to the client.
	// The baseDir is the directory where the files will be downloaded to.
	// Should not include specific location, just the base directory. The torrent hash will be appended to it.
	// Returns the torrent that was added
	AddDownload(infoHash string, baseDir string) (*Torrent, error)

	// Adds a new download from a url to the client.
	// The baseDir is the directory where the files will be downloaded to.
	// Should not include specific location, just the base directory. The torrent hash will be appended to it.
	// Returns the torrent that was added
	AddDownloadFromUrl(url string, baseDir string) (*Torrent, error)

	// Removes a download from the wrapper, optionally deleting the files
	RemoveDownload(infoHash string, deleteFiles bool) error

	// Returns a map of all running torrents, indexed by their info hash
	GetRunningTorrents() map[string]*Torrent

	GetBaseDir() string
}

func NewTorrent(t *torrent.Torrent, baseDir string) *Torrent {
	return &Torrent{
		t:       t,
		key:     t.InfoHash().HexString(),
		baseDir: baseDir,
		lastSpeed: SpeedData{
			t:     time.Now(),
			bytes: 0,
		},
	}
}

// Returns the wrapped torrent.Torrent
func (t *Torrent) GetTorrent() *torrent.Torrent {
	return t.t
}

// Returns useful information about the torrent
func (t *Torrent) GetInfo() TorrentInfo {
	c := t.t.Stats().BytesReadData
	bytesRead := c.Int64()
	var speed int64 = 0

	bytesDiff := bytesRead - t.lastSpeed.bytes
	timeDiff := time.Since(t.lastSpeed.t).Seconds()
	speed = int64(float64(bytesDiff) / timeDiff)

	t.lastSpeed = SpeedData{
		t:     time.Now(),
		bytes: bytesRead,
	}

	return TorrentInfo{
		InfoHash:  t.key,
		Name:      t.t.Name(),
		Size:      t.t.Length(),
		Progress:  t.t.BytesCompleted(),
		Completed: percent(t.t.BytesCompleted(), t.t.Length()),
		Speed:     humanReadableSpeed(speed),
	}
}

func humanReadableSpeed(s int64) string {
	speed := float64(s)
	if speed < 1024 {
		return fmt.Sprintf("%.2f B/s", speed)
	}
	speed /= 1024
	if speed < 1024 {
		return fmt.Sprintf("%.2f KB/s", speed)
	}
	speed /= 1024
	return fmt.Sprintf("%.2f MB/s", speed)
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
