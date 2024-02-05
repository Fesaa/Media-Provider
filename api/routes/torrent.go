package routes

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

var lock *sync.RWMutex = &sync.RWMutex{}

type SpeedData struct {
	t     time.Time
	bytes int64
}

var speedMap map[string]SpeedData = make(map[string]SpeedData)

func Download(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not found in context")
		return fiber.ErrInternalServerError
	}

	torrentProvider := holder.GetTorrentProvider()
	if torrentProvider == nil {
		slog.Error("Torrent provider not found in holder")
		return fiber.ErrInternalServerError
	}

	infoHash := ctx.Params("infoHash")
	if infoHash == "" {
		slog.Error("No infoHash provided")
		return fiber.ErrBadRequest
	}

	_, err := torrentProvider.AddDownload(infoHash)
	if err != nil {
		slog.Error(err.Error())
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stop(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not found in context")
		return fiber.ErrInternalServerError
	}

	torrentProvider := holder.GetTorrentProvider()
	if torrentProvider == nil {
		slog.Error("Torrent provider not found in holder")
		return fiber.ErrInternalServerError
	}

	infoHash := ctx.Params("infoHash")
	if infoHash == "" {
		slog.Error("No infoHash provided")
		return fiber.ErrBadRequest
	}

	err := torrentProvider.RemoveDownload(infoHash)
	if err != nil {
		slog.Error(err.Error())
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stats(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not found in context")
		return fiber.ErrInternalServerError
	}

	torrentProvider := holder.GetTorrentProvider()
	if torrentProvider == nil {
		slog.Error("Torrent provider not found in holder")
		return fiber.ErrInternalServerError
	}

	torrents := torrentProvider.GetRunningTorrents()
	info := make(map[string]interface{})
	for key, torrent := range torrents {
		c := torrent.Stats().BytesReadData
		progress := c.Int64()
		var speed int64 = 0

		lock.RLock()
		s, ok := speedMap[key]
		lock.RUnlock()
		if ok {
			bytesDiff := progress - s.bytes
			timeDiff := time.Now().Sub(s.t).Seconds()
			speed = int64(float64(bytesDiff) / timeDiff)
		}

		lock.Lock()
		speedMap[key] = SpeedData{
			t:     time.Now(),
			bytes: progress,
		}
		lock.Unlock()

		info[key] = map[string]interface{}{
			"InfoHash":  torrent.InfoHash().HexString(),
			"Name":      torrent.Name(),
			"Size":      torrent.Length(),
			"Progress":  progress,
			"Completed": percent(progress, torrent.Length()),
			"Speed":     humanReadableSpeed(speed),
		}
	}
	return ctx.JSON(info)
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
