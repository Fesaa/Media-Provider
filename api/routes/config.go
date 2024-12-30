package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"strconv"
)

type configRoutes struct {
	dig.In

	Cfg    *config.Config
	Router fiber.Router
	Auth   auth.Provider `name:"jwt-auth"`
	Log    zerolog.Logger
	Val    *validator.Validate
}

func RegisterConfigRoutes(cr configRoutes) {
	configGroup := cr.Router.Group("/config", cr.Auth.Middleware)
	configGroup.Get("/", cr.GetConfig)
	configGroup.Post("/update", cr.UpdateConfig)
}

func (cr *configRoutes) GetConfig(ctx *fiber.Ctx) error {
	cp := *cr.Cfg
	cp.Secret = ""
	return ctx.JSON(cp)
}

func (cr *configRoutes) UpdateConfig(ctx *fiber.Ctx) error {
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		cr.Log.Debug().Err(err).Msg("invalid sync id")
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	var c config.Config
	if err = ctx.BodyParser(&c); err != nil {
		cr.Log.Debug().Err(err).Msg("invalid config body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config"})
	}

	if err = cr.Val.Struct(c); err != nil {
		cr.Log.Debug().Err(err).Msg("invalid config")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err = cr.Cfg.Update(c, syncID); err != nil {
		cr.Log.Error().Err(err).Msg("failed to update config")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(cr.Cfg.SyncId))
}
