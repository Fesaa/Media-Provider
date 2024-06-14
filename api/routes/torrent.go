package routes

import (
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/yoitsu"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func Download(ctx *fiber.Ctx) error {
	var req providers.DownloadRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		slog.Error("Error parsing request body into DownloadRequest", "err", err)
		return fiber.ErrBadRequest
	}

	if req.BaseDir == "" {
		slog.Warn("Trying to download Torrent to empty baseDir, returning error.")
		return fiber.ErrBadRequest
	}

	err = providers.Download(req)
	if err != nil {
		slog.Error("Error adding download", "error", err, "debug_info", req.DebugString())
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stop(ctx *fiber.Ctx) error {
	infoHash := ctx.Params("infoHash")
	if infoHash == "" {
		slog.Error("No infoHash provided")
		return fiber.ErrBadRequest
	}

	err := yoitsu.I().RemoveDownload(infoHash, true)
	if err != nil {
		slog.Error("Error stopping download", "infoHash", infoHash, "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stats(ctx *fiber.Ctx) error {
	torrents := yoitsu.I().GetRunningTorrents()
	info := make(map[string]yoitsu.TorrentInfo, torrents.Len())
	torrents.ForEachSafe(func(key string, torrent yoitsu.Torrent) {
		info[key] = torrent.GetInfo()
	})
	return ctx.JSON(info)
}
