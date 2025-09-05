package routes

import (
	"errors"
	"fmt"

	"github.com/Fesaa/Media-Provider/db/models"
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
	Cache  fiber.Handler `name:"cache"`
	Auth   services.AuthService
	YS     yoitsu.Client
	PS     core.Client
	Log    zerolog.Logger

	Val            services.ValidationService
	ContentService services.ContentService
	Transloco      services.TranslocoService
}

func RegisterContentRoutes(cr contentRoutes) {
	cr.Router.Group("/content", cr.Auth.Middleware).
		Post("/search", cr.Cache, withBodyValidation(cr.Search)).
		Post("/download", withBodyValidation(cr.Download)).
		Post("/stop", withBodyValidation(cr.Stop)).
		Get("/stats", withParam(newQueryParam("all", withAllowEmpty(false)), cr.Stats)).
		Post("/message", withBody(cr.Message))
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
	user := services.GetFromContext(ctx, services.UserKey)

	if req.BaseDir == "" {
		cr.Log.Warn().Msg("trying to download Torrent to empty baseDir, returning error.")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": cr.Transloco.GetTranslation("base-dir-not-empty"),
		})
	}

	req.OwnerId = user.ID

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

func (cr *contentRoutes) Stats(ctx *fiber.Ctx, allDownloads bool) error {
	user := services.GetFromContext(ctx, services.UserKey)
	allDownloads = allDownloads && user.HasRole(models.ViewAllDownloads)

	statsResponse := payload.StatsResponse{
		Running:      []payload.InfoStat{},
		TotalRunning: map[models.Provider]int{},
	}
	cr.YS.GetTorrents().ForEachSafe(func(_ string, torrent yoitsu.Torrent) {
		if allDownloads || torrent.Request().OwnerId == user.ID {
			statsResponse.Running = append(statsResponse.Running, torrent.GetInfo())
		}

		statsResponse.TotalRunning[torrent.Provider()]++
	})
	for _, download := range cr.PS.GetCurrentDownloads() {
		if allDownloads || download.Request().OwnerId == user.ID {
			statsResponse.Running = append(statsResponse.Running, download.GetInfo())
		}

		statsResponse.TotalRunning[download.Provider()]++
	}

	return ctx.JSON(statsResponse)
}
