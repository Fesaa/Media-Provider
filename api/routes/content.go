package routes

import (
	"errors"
	"fmt"

	"github.com/Fesaa/Media-Provider/api/middleware"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type contentRoutes struct {
	dig.In

	Router fiber.Router
	Cache  fiber.Handler        `name:"cache"`
	Auth   services.AuthService `name:"jwt-auth"`
	YS     yoitsu.Client
	PS     core.Client
	Log    zerolog.Logger

	Val            services.ValidationService
	ContentService services.ContentService
	Transloco      services.TranslocoService
}

func RegisterContentRoutes(cr contentRoutes) {
	router := cr.Router.Group("/content", cr.Auth.Middleware)
	router.Post("/search", cr.Cache, middleware.WithBodyValidation(cr.Search))
	router.Post("/download", middleware.WithBodyValidation(cr.Download))
	router.Post("/stop", middleware.WithBodyValidation(cr.Stop))
	router.Get("/stats", cr.Stats)
	router.Post("/message", middleware.WithBody(cr.Message))
}

func (cr *contentRoutes) Message(ctx *fiber.Ctx, msg payload.Message) error {
	resp, err := cr.ContentService.Message(msg)
	if err != nil {
		if errors.Is(err, services.ErrContentNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if errors.Is(err, services.ErrUnknownMessageType) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		if errors.Is(err, services.ErrWrongState) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		cr.Log.Error().Err(err).Msg("An error occurred while sending a message down to Content")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(resp)
}

func (cr *contentRoutes) Search(ctx *fiber.Ctx, searchRequest payload.SearchRequest) error {
	search, err := cr.ContentService.Search(searchRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if len(search) == 0 {
		return ctx.JSON([]payload.Info{})
	}

	return ctx.JSON(search)
}

func (cr *contentRoutes) Download(ctx *fiber.Ctx, req payload.DownloadRequest) error {
	if req.BaseDir == "" {
		cr.Log.Warn().Msg("trying to download Torrent to empty baseDir, returning error.")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": cr.Transloco.GetTranslation("base-dir-not-empty"),
		})
	}

	if err := cr.ContentService.Download(req); err != nil {
		cr.Log.Error().
			Err(err).
			Str("debug_info", fmt.Sprintf("%#v", req)).
			Msg("error while downloading torrent")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (cr *contentRoutes) Stop(ctx *fiber.Ctx, req payload.StopRequest) error {
	if err := cr.ContentService.Stop(req); err != nil {
		cr.Log.Error().Str("id", req.Id).Msg("error while stopping download")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (cr *contentRoutes) Stats(ctx *fiber.Ctx) error {
	statsResponse := payload.StatsResponse{
		Running: []payload.InfoStat{},
	}
	cr.YS.GetTorrents().ForEachSafe(func(_ string, torrent yoitsu.Torrent) {
		statsResponse.Running = append(statsResponse.Running, torrent.GetInfo())
	})
	for _, download := range cr.PS.GetCurrentDownloads() {
		statsResponse.Running = append(statsResponse.Running, download.GetInfo())
	}

	return ctx.JSON(statsResponse)
}
