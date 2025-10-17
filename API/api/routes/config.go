package routes

import (
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type configRoutes struct {
	dig.In

	Cfg                 *config.Config
	Router              fiber.Router
	Auth                services.AuthService
	Val                 services.ValidationService
	SettingsService     services.SettingsService
	SubscriptionService services.SubscriptionService
	SignalR             services.SignalRService
	TransLoco           services.TranslocoService
}

func RegisterConfigRoutes(cr configRoutes) {
	cr.Router.Group("/config").
		Get("/oidc", cr.getOidcConfig).
		Use(cr.Auth.Middleware).
		Get("/", cr.getConfig).
		Post("/", withParams(cr.updateConfig, newBodyParam[payload.Settings]()))
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

func (cr *configRoutes) updateConfig(ctx *fiber.Ctx, dto payload.Settings) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	cur, err := cr.SettingsService.GetSettingsDto(ctx.Context())
	if err != nil {
		return err
	}

	if err = cr.SettingsService.UpdateSettingsDto(ctx.UserContext(), dto); err != nil {
		log.Error().Err(err).Msg("failed to update config")
		return InternalError(err)
	}

	if cur.SubscriptionRefreshHour != dto.SubscriptionRefreshHour {
		if err = cr.SubscriptionService.UpdateTask(ctx.UserContext(), dto.SubscriptionRefreshHour); err != nil {
			log.Error().Err(err).Msg("failed to update subscription refresh hour")
			cr.SignalR.Notify(ctx.UserContext(), models.NewNotification().
				WithTitle(cr.TransLoco.GetTranslation("failed-to-register-sub-task-title")).
				WithSummary(cr.TransLoco.GetTranslation("failed-to-register-sub-task-summary")).
				WithOwner(contextkey.GetFromCtxOrDefault(ctx.Context(), contextkey.User).ID).
				WithGroup(models.GroupError).
				WithColour(models.Error).
				Build())
		} else {
			log.Info().Int("hour", dto.SubscriptionRefreshHour).Msg("Subscription hour has been updated")
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}
