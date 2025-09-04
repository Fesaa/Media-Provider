package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type configRoutes struct {
	dig.In

	Cfg             *config.Config
	Router          fiber.Router
	Auth            services.AuthService `name:"jwt-auth"`
	Val             services.ValidationService
	SettingsService services.SettingsService
	Log             zerolog.Logger
}

func RegisterConfigRoutes(cr configRoutes) {
	configGroup := cr.Router.Group("/config")

	// Auth
	configGroup.Get("/", cr.Auth.Middleware, cr.getConfig)
	configGroup.Post("/", cr.Auth.Middleware, withBody(cr.updateConfig))

	// No Auth
	configGroup.Get("/oidc", cr.getOidcConfig)
}

func (cr *configRoutes) getConfig(ctx *fiber.Ctx) error {
	dto, err := cr.SettingsService.GetSettingsDto()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return ctx.JSON(dto)
}

func (cr *configRoutes) getOidcConfig(ctx *fiber.Ctx) error {
	dto, err := cr.SettingsService.GetSettingsDto()
	if err != nil {
		cr.Log.Error().Err(err).Msg("Failed to get oidc config")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
		})
	}
	return ctx.JSON(dto.Oidc)
}

func (cr *configRoutes) updateConfig(ctx *fiber.Ctx, c payload.Settings) error {
	if err := cr.SettingsService.UpdateSettingsDto(c); err != nil {
		cr.Log.Error().Err(err).Msg("failed to update config")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}
