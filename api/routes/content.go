package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/gofiber/fiber/v2"
)

type contentRoutes struct{}

func RegisterContentRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	cr := contentRoutes{}
	router.Post("/search", auth.Middleware, cache, wrap(cr.Search))
	router.Post("/download", auth.Middleware, wrap(cr.Download))
	router.Post("/stop", auth.Middleware, wrap(cr.Stop))
	router.Get("/stats", auth.Middleware, wrap(cr.Stats))
}

func (cr *contentRoutes) Download(l *log.Logger, ctx *fiber.Ctx) error {
	var req payload.DownloadRequest
	if err := ctx.BodyParser(&req); err != nil {
		l.Error("error while parsing request body into DownloadRequest", "err", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.BaseDir == "" {
		l.Warn("trying to download Torrent to empty baseDir, returning error.")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "base dir cannot be null",
		})
	}

	if err := providers.Download(req); err != nil {
		l.Error("error while adding download", "error", err, "debug_info", fmt.Sprintf("%#v", req))
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (cr *contentRoutes) Stop(l *log.Logger, ctx *fiber.Ctx) error {
	var req payload.StopRequest
	if err := ctx.BodyParser(&req); err != nil {
		l.Error("error while parsing request body into StopRequest", "err", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := providers.Stop(req); err != nil {
		l.Error("error while stopping download", "id", req.Id, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (cr *contentRoutes) Stats(l *log.Logger, ctx *fiber.Ctx) error {
	statsResponse := payload.StatsResponse{
		Running: []payload.InfoStat{},
		Queued:  []payload.QueueStat{},
	}
	yoitsu.I().GetRunningTorrents().ForEachSafe(func(key string, torrent yoitsu.Torrent) {
		statsResponse.Running = append(statsResponse.Running, torrent.GetInfo())
	})
	for _, download := range pasloe.I().GetCurrentDownloads() {
		statsResponse.Running = append(statsResponse.Running, download.GetInfo())
	}

	for _, queueStat := range yoitsu.I().GetQueuedTorrents() {
		statsResponse.Queued = append(statsResponse.Queued, queueStat)
	}
	for _, queueStat := range pasloe.I().GetQueuedDownloads() {
		statsResponse.Queued = append(statsResponse.Queued, queueStat)
	}
	return ctx.JSON(statsResponse)
}
