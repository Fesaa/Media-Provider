package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type preferencesRoute struct {
	dig.In

	Router fiber.Router
	Auth   auth.Provider `name:"jwt-auth"`
	DB     *db.Database
	Log    zerolog.Logger
	Val    *validator.Validate
}

func RegisterPreferencesRoutes(pr preferencesRoute) {
	group := pr.Router.Group("/preferences", pr.Auth.Middleware)

	group.Get("/", pr.Get)
	group.Post("/save", pr.Update)
}

func (pr *preferencesRoute) Get(ctx *fiber.Ctx) error {
	pref, err := pr.DB.Preferences.Get()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get preferences")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(pref)
}

func (pr *preferencesRoute) Update(ctx *fiber.Ctx) error {
	var pref models.Preference
	if err := ctx.BodyParser(&pref); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to parse preferences")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := pr.Val.Struct(&pref); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to validate preferences")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := pr.DB.Preferences.Update(pref); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to update preferences")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
