package yoitsu

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
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
	t        *torrent.Torrent
	key      string
	baseDir  string
	provider config.Provider

	ctx       context.Context
	cancel    context.CancelFunc
	lastSpeed SpeedData
}

func newTorrent(t *torrent.Torrent, baseDir string, provider config.Provider) Torrent {
	return &torrentImpl{
		t:        t,
		key:      t.InfoHash().HexString(),
		baseDir:  baseDir,
		provider: provider,
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

func (t *torrentImpl) GetInfo() config.Info {
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

	return config.Info{
		Provider:  t.provider,
		InfoHash:  t.key,
		Name:      t.t.Name(),
		Size:      utils.BytesToSize(float64(t.t.Length())),
		Progress:  t.t.BytesCompleted(),
		Completed: utils.Percent(t.t.BytesCompleted(), t.t.Length()),
		Speed:     utils.HumanReadableSpeed(speed),
	}
}
