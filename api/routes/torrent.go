package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/yoitsu"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func Download(ctx *fiber.Ctx) error {
	var req payload.DownloadRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.Error("Error parsing request body into DownloadRequest", "err", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.BaseDir == "" {
		slog.Warn("Trying to download Torrent to empty baseDir, returning error.")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "base dir cannot be null",
		})
	}

	if err := providers.Download(req); err != nil {
		slog.Error("Error adding download", "error", err, "debug_info", fmt.Sprintf("%#v", req))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stop(ctx *fiber.Ctx) error {
	var req payload.StopRequest
	if err := ctx.BodyParser(&req); err != nil {
		slog.Error("Error parsing request body into StopRequest", "err", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := providers.Stop(req); err != nil {
		slog.Error("Error stopping download", "id", req.Id, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stats(ctx *fiber.Ctx) error {
	statsResponse := payload.StatsResponse{
		Running: []payload.InfoStat{},
		Queued:  []payload.QueueStat{},
	}
	yoitsu.I().GetRunningTorrents().ForEachSafe(func(key string, torrent yoitsu.Torrent) {
		statsResponse.Running = append(statsResponse.Running, torrent.GetInfo())
	})
	manga := mangadex.I().GetCurrentManga()
	if manga != nil {
		statsResponse.Running = append(statsResponse.Running, manga.GetInfo())
	}
	for _, id := range yoitsu.I().GetQueuedTorrents() {
		statsResponse.Queued = append(statsResponse.Queued, id)
	}
	for _, id := range mangadex.I().GetQueuedMangas() {
		statsResponse.Queued = append(statsResponse.Queued, id)
	}
	return ctx.JSON(statsResponse)
}
