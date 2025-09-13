package routes

import (
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type configRoutes struct {
	dig.In

	Cfg             *config.Config
	Router          fiber.Router
	Auth            services.AuthService
	Val             services.ValidationService
	SettingsService services.SettingsService
}

func RegisterConfigRoutes(cr configRoutes) {
	cr.Router.Group("/config").
		Get("/oidc", cr.getOidcConfig).
		Use(cr.Auth.Middleware).
		Get("/", cr.getConfig).
		Post("/", withBody(cr.updateConfig))
}

func (cr *configRoutes) getConfig(ctx *fiber.Ctx) error {
	dto, err := cr.SettingsService.GetSettingsDto(ctx.UserContext())
	if err != nil {
		return InternalError(err)
	}

	if dto.Oidc.ClientSecret != "" {
		dto.Oidc.ClientSecret = strings.Repeat("*", len(dto.Oidc.ClientSecret))
	}

	return ctx.JSON(dto)
}

func (cr *configRoutes) getOidcConfig(ctx *fiber.Ctx) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	dto, err := cr.SettingsService.GetSettingsDto(ctx.UserContext())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get oidc config")
		return InternalError(err)
	}
	return ctx.JSON(payload.PublicOidcSettings{
		DisablePasswordLogin: dto.Oidc.DisablePasswordLogin,
		AutoLogin:            dto.Oidc.AutoLogin,
		Enabled:              dto.Oidc.Enabled(),
	})
}

func (cr *configRoutes) updateConfig(ctx *fiber.Ctx, settings payload.Settings) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	if err := cr.SettingsService.UpdateSettingsDto(ctx.UserContext(), settings); err != nil {
		log.Error().Err(err).Msg("failed to update config")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}
