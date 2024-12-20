package yoitsu

import (
	"context"
	"fmt"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/anacrolix/torrent"
	"log/slog"
	"path"
	"time"
)

// torrentImpl wrapper around the torrent.Torrent struct
// Providers some specific functionality
type torrentImpl struct {
	t   *torrent.Torrent
	log *log.Logger

	key       string
	baseDir   string
	tempTitle string
	provider  models.Provider

	ctx    context.Context
	cancel context.CancelFunc

	lastTime time.Time
	lastRead int64
}

func newTorrent(t *torrent.Torrent, req payload.DownloadRequest) Torrent {
	tor := &torrentImpl{
		t:         t,
		key:       t.InfoHash().HexString(),
		baseDir:   req.BaseDir,
		tempTitle: req.TempTitle,
		provider:  req.Provider,
		lastTime:  time.Now(),
		lastRead:  0,
	}

	tor.log = log.With(slog.String("infoHash", tor.key))
	return tor
}

func (t *torrentImpl) GetTorrent() *torrent.Torrent {
	return t.t
}

func (t *torrentImpl) WaitForInfoAndDownload() {
	if t.cancel != nil {
		t.log.Debug("already loading info")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	t.log.Trace("loading torrent info")
	go func() {
		select {
		case <-t.ctx.Done():
			return
		case <-t.t.GotInfo():
			t.log = t.log.With("name", t.t.Info().BestName())
			t.log.Debug("starting torrent download")
			t.t.DownloadAll()
		}
	}()
}

func (t *torrentImpl) Cancel() error {
	t.log.Trace("canceling torrent")
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
		Estimated: func() *int64 {
			if speed == 0 {
				return nil
			}
			es := (t.t.Length() - bytesRead) / speed
			return &es
		}(),
		SpeedType:   payload.BYTES,
		Speed:       payload.SpeedData{T: time.Now().Unix(), Speed: speed},
		DownloadDir: t.GetDownloadDir(),
	}
}
