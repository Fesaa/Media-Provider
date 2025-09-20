package routes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/providers/pasloe/core"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type contentRoutes struct {
	dig.In

	Router fiber.Router
	Cache  fiber.Handler `name:"cache"`
	Auth   services.AuthService
	YS     yoitsu.Client
	PS     core.Client

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
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	resp, err := cr.ContentService.Message(msg)
	if err != nil {
		if errors.Is(err, services.ErrContentNotFound) {
			return NotFound(err)
		}

		if errors.Is(err, services.ErrUnknownMessageType) {
			return BadRequest(err)
		}

		if errors.Is(err, services.ErrWrongState) {
			return BadRequest(err)
		}

		log.Error().Err(err).Msg("An error occurred while sending a message down to Content")
		return InternalError(err)
	}

	return ctx.JSON(resp)
}

func (cr *contentRoutes) Search(ctx *fiber.Ctx, searchRequest payload.SearchRequest) error {
	search, err := cr.ContentService.Search(ctx.UserContext(), searchRequest)
	if err != nil {
		return InternalError(err)
	}

	if len(search) == 0 {
		return ctx.JSON([]payload.Info{})
	}

	return ctx.JSON(search)
}

func (cr *contentRoutes) Download(ctx *fiber.Ctx, req payload.DownloadRequest) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)

	if req.BaseDir == "" {
		log.Warn().Msg("trying to download to empty baseDir, returning error.")
		return BadRequest(errors.New(cr.Transloco.GetTranslation("base-dir-not-empty")))
	}

	if strings.Contains(req.BaseDir, "..") {
		return BadRequest()
	}

	req.OwnerId = user.ID

	if err := cr.ContentService.Download(req); err != nil {
		log.Error().
			Err(err).
			Str("debug_info", fmt.Sprintf("%#v", req)).
			Msg("error while downloading")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (cr *contentRoutes) Stop(ctx *fiber.Ctx, req payload.StopRequest) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	if err := cr.ContentService.Stop(req); err != nil {
		log.Error().Err(err).Str("id", req.Id).Msg("error while stopping download")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (cr *contentRoutes) Stats(ctx *fiber.Ctx, allDownloads bool) error {
	user := contextkey.GetFromContext(ctx, contextkey.User)
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
