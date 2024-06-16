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
		slog.Error("Error adding download", "error", err, "debug_info", fmt.Sprintf("%#v", req))
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func Stop(ctx *fiber.Ctx) error {
	var req payload.StopRequest
	err := ctx.BodyParser(&req)
	if err != nil {
		slog.Error("Error parsing request body into StopRequest", "err", err)
		return fiber.ErrBadRequest
	}
	id := ctx.Params("id")

	err = providers.Stop(req)
	if err != nil {
		slog.Error("Error stopping download", "id", id, "error", err)
		return fiber.ErrInternalServerError
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
