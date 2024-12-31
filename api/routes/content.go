package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type contentRoutes struct {
	dig.In

	Router   fiber.Router
	Cache    fiber.Handler `name:"cache"`
	Auth     auth.Provider `name:"jwt-auth"`
	Provider *providers.ContentProvider
	YS       yoitsu.Yoitsu
	PS       api.Client
	Log      zerolog.Logger
}

func RegisterContentRoutes(cr contentRoutes) {
	cr.Router.Post("/search", cr.Auth.Middleware, cr.Cache, cr.Search)
	cr.Router.Post("/download", cr.Auth.Middleware, cr.Download)
	cr.Router.Post("/stop", cr.Auth.Middleware, cr.Stop)
	cr.Router.Get("/stats", cr.Auth.Middleware, cr.Stats)
}

func (cr *contentRoutes) Search(ctx *fiber.Ctx) error {
	var searchRequest payload.SearchRequest
	if err := ctx.BodyParser(&searchRequest); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	search, err := cr.Provider.Search(searchRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(search)
}

func (cr *contentRoutes) Download(ctx *fiber.Ctx) error {
	var req payload.DownloadRequest
	if err := ctx.BodyParser(&req); err != nil {
		cr.Log.Error().Err(err).Msg("error while parsing body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.BaseDir == "" {
		cr.Log.Warn().Msg("trying to download Torrent to empty baseDir, returning error.")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "base dir cannot be null",
		})
	}

	if err := cr.Provider.Download(req); err != nil {
		cr.Log.Error().
			Err(err).
			Str("debug_info", fmt.Sprintf("%#v", req)).
			Msg("error while downloading torrent")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (cr *contentRoutes) Stop(ctx *fiber.Ctx) error {
	var req payload.StopRequest
	if err := ctx.BodyParser(&req); err != nil {
		cr.Log.Error().Err(err).Msg("error while parsing body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := cr.Provider.Stop(req); err != nil {
		cr.Log.Error().Str("id", req.Id).Msg("error while stopping download")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusAccepted)
}

func (cr *contentRoutes) Stats(ctx *fiber.Ctx) error {
	statsResponse := payload.StatsResponse{
		Running: []payload.InfoStat{},
		Queued:  []payload.QueueStat{},
	}
	cr.YS.GetRunningTorrents().ForEachSafe(func(_ string, torrent yoitsu.Torrent) {
		statsResponse.Running = append(statsResponse.Running, torrent.GetInfo())
	})
	for _, download := range cr.PS.GetCurrentDownloads() {
		statsResponse.Running = append(statsResponse.Running, download.GetInfo())
	}

	statsResponse.Queued = append(statsResponse.Queued, cr.YS.GetQueuedTorrents()...)
	statsResponse.Queued = append(statsResponse.Queued, cr.PS.GetQueuedDownloads()...)

	return ctx.JSON(statsResponse)
}
