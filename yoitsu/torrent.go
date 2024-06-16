package yoitsu

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"log/slog"
	"path"
	"time"
)

// torrentImpl wrapper around the torrent.Torrent struct
// Providers some specific functionality
type torrentImpl struct {
	t         *torrent.Torrent
	key       string
	baseDir   string
	tempTitle string
	provider  config.Provider

	ctx    context.Context
	cancel context.CancelFunc

	lastTime time.Time
	lastRead int64
}

func newTorrent(t *torrent.Torrent, req payload.DownloadRequest) Torrent {
	return &torrentImpl{
		t:         t,
		key:       t.InfoHash().HexString(),
		baseDir:   req.BaseDir,
		tempTitle: req.TempTitle,
		provider:  req.Provider,
		lastTime:  time.Now(),
		lastRead:  0,
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

func (t *torrentImpl) GetDownloadDir() string {
	return path.Join(t.baseDir, t.key)
}

func (t *torrentImpl) GetInfo() payload.InfoStat {
	c := t.t.Stats().BytesReadData
	bytesRead := c.Int64()
	bytesDiff := bytesRead - t.lastRead
	timeDiff := max(time.Since(t.lastTime).Seconds(), 1)
	speed := int64(float64(bytesDiff) / timeDiff)
	t.lastRead = bytesRead
	t.lastTime = time.Now()

	return payload.InfoStat{
		Provider: t.provider,
		Id:       t.key,
		Name: func() string {
			if t.t.Info() != nil {
				return t.t.Info().BestName()
			}
			return t.tempTitle
		}(),
		Size:        utils.BytesToSize(float64(t.t.Length())),
		Downloading: t.t.Info() != nil,
		Progress:    utils.Percent(t.t.BytesCompleted(), t.t.Length()),
		SpeedType:   payload.BYTES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: t.GetDownloadDir(),
	}
}
