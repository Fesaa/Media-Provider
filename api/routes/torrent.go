package routes

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

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
	for key, value := range torrents {
		info[key] = map[string]interface{}{
			"InfoHash":  value.InfoHash().HexString(),
			"Name":      value.Name(),
			"Size":      value.Length(),
			"Progress":  value.BytesCompleted(),
			"Completed": percent(value.BytesCompleted(), value.Length()),
		}
	}
	return ctx.JSON(info)
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
