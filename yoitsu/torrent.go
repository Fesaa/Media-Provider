package yoitsu

import (
	"context"
	"fmt"
	"github.com/anacrolix/torrent"
	"log/slog"
	"time"
)

// SpeedData contains the amount of download bytes at a certain time
type SpeedData struct {
	t     time.Time
	bytes int64
}

// torrentImpl wrapper around the torrent.Torrent struct
// Providers some specific functionality
type torrentImpl struct {
	t       *torrent.Torrent
	key     string
	baseDir string

	ctx       context.Context
	cancel    context.CancelFunc
	lastSpeed SpeedData
}

func newTorrent(t *torrent.Torrent, baseDir string) Torrent {
	return &torrentImpl{
		t:       t,
		key:     t.InfoHash().HexString(),
		baseDir: baseDir,
		lastSpeed: SpeedData{
			t:     time.Now(),
			bytes: 0,
		},
	}
}

func (t *torrentImpl) GetTorrent() *torrent.Torrent {
	return t.t
}

func (t *torrentImpl) WaitForInfoAndDownload() {
	if t.cancel != nil {
		slog.Debug("Yoitsu has already started loading info", "name", t.t.Name(), "infoHash", t.t.InfoHash().HexString())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	slog.Debug("Starting loading Torrent info", "infoHash", t.t.InfoHash().HexString())
	go func() {
		select {
		case <-t.ctx.Done():
			return
		case <-t.t.GotInfo():
			slog.Info("Starting download", "name", t.t.Name(), "infoHash", t.t.InfoHash().HexString())
			t.t.DownloadAll()
		}
	}()
}

func (t *torrentImpl) Cancel() error {
	if t.cancel == nil {
		return fmt.Errorf("torrent is not downloading")
	}
	t.cancel()
	return nil
}

func (t *torrentImpl) GetInfo() TorrentInfo {
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