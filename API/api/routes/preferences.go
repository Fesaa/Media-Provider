package routes

import (
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type preferencesRoute struct {
	dig.In

	Router     fiber.Router
	Auth       services.AuthService
	Val        services.ValidationService
	UnitOfWork *db.UnitOfWork
}

func RegisterPreferencesRoutes(pr preferencesRoute) {
	pr.Router.Group("/preferences", pr.Auth.Middleware).
		Get("/", pr.get).
		Post("/save", withBody(pr.update))
}

func (pr *preferencesRoute) get(ctx *fiber.Ctx) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)

	pref, err := pr.UnitOfWork.Preferences.GetPreferences(ctx.UserContext(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get preferences")
		return InternalError(err)
	}
	return ctx.JSON(pref)
}

func (pr *preferencesRoute) update(ctx *fiber.Ctx, pref models.UserPreferences) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)

	pref.UserID = user.ID
	if err := pr.UnitOfWork.Preferences.Update(ctx.UserContext(), &pref); err != nil {
		log.Error().Err(err).Msg("Failed to update preferences")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
