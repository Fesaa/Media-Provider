package yoitsu

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
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
		log.Debug("Yoitsu has already started loading info", "infoHash", t.key, "name", t.t.Name())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	log.Trace("loading torrent info", "infoHash", t.key)
	go func() {
		select {
		case <-t.ctx.Done():
			return
		case <-t.t.GotInfo():
			log.Debug("starting torrent download", "infoHash", t.key, "name", t.t.Name())
			t.t.DownloadAll()
		}
	}()
}

func (t *torrentImpl) Cancel() error {
	log.Trace("calling cancel on torrent", "infoHash", t.key)
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
