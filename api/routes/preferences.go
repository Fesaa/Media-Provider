package routes

import (
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type preferencesRoute struct {
	dig.In

	Router fiber.Router
	Auth   services.AuthService
	DB     *db.Database
	Val    services.ValidationService
	Pref   services.PreferencesService
}

func RegisterPreferencesRoutes(pr preferencesRoute) {
	pr.Router.Group("/preferences", pr.Auth.Middleware).
		Get("/", pr.get).
		Post("/save", hasRole(models.ManagePreferences), withBody(pr.update))
}

func (pr *preferencesRoute) get(ctx *fiber.Ctx) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	pref, err := pr.Pref.GetDto()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get preferences")
		return InternalError(err)
	}
	return ctx.JSON(pref)
}

func (pr *preferencesRoute) update(ctx *fiber.Ctx, pref payload.PreferencesDto) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	if err := pr.Pref.Update(pref); err != nil {
		log.Error().Err(err).Msg("Failed to update preferences")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
