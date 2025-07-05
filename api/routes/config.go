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
	configGroup.Get("/", cr.Auth.Middleware, cr.GetConfig)
	configGroup.Post("/", cr.Auth.Middleware, cr.UpdateConfig)

	// No Auth
	configGroup.Get("/oidc", cr.GetOidcConfig)
}

func (cr *configRoutes) GetConfig(ctx *fiber.Ctx) error {
	dto, err := cr.SettingsService.GetSettingsDto()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return ctx.JSON(dto)
}

func (cr *configRoutes) GetOidcConfig(ctx *fiber.Ctx) error {
	dto, err := cr.SettingsService.GetSettingsDto()
	if err != nil {
		cr.Log.Error().Err(err).Msg("Failed to get oidc config")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
		})
	}
	return ctx.JSON(dto.Oidc)
}

func (cr *configRoutes) UpdateConfig(ctx *fiber.Ctx) error {
	var c payload.Settings
	if err := cr.Val.ValidateCtx(ctx, &c); err != nil {
		cr.Log.Debug().Err(err).Msg("invalid config")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := cr.SettingsService.UpdateSettingsDto(c); err != nil {
		cr.Log.Error().Err(err).Msg("failed to update config")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
