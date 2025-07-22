package routes

import (
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type preferencesRoute struct {
	dig.In

	Router fiber.Router
	Auth   services.AuthService `name:"jwt-auth"`
	DB     *db.Database
	Log    zerolog.Logger
	Val    services.ValidationService
	Pref   services.PreferencesService
}

func RegisterPreferencesRoutes(pr preferencesRoute) {
	group := pr.Router.Group("/preferences", pr.Auth.Middleware)

	group.Get("/", pr.Get)
	group.Post("/save", pr.Update)
}

func (pr *preferencesRoute) Get(ctx *fiber.Ctx) error {
	pref, err := pr.Pref.GetDto()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get preferences")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.JSON(pref)
}

func (pr *preferencesRoute) Update(ctx *fiber.Ctx) error {
	var pref payload.PreferencesDto
	if err := pr.Val.ValidateCtx(ctx, &pref); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to parse preferences")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := pr.Pref.Update(pref); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to update preferences")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
