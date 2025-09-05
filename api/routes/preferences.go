package routes

import (
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type preferencesRoute struct {
	dig.In

	Router fiber.Router
	Auth   services.AuthService
	DB     *db.Database
	Log    zerolog.Logger
	Val    services.ValidationService
	Pref   services.PreferencesService
}

func RegisterPreferencesRoutes(pr preferencesRoute) {
	pr.Router.Group("/preferences", pr.Auth.Middleware).
		Get("/", pr.get).
		Post("/save", hasRole(models.ManagePreferences), withBody(pr.update))
}

func (pr *preferencesRoute) get(ctx *fiber.Ctx) error {
	pref, err := pr.Pref.GetDto()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get preferences")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.JSON(pref)
}

func (pr *preferencesRoute) update(ctx *fiber.Ctx, pref payload.PreferencesDto) error {
	if err := pr.Pref.Update(pref); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to update preferences")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
