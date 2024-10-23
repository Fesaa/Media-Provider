package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

type configRoutes struct{}

func RegisterConfigRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	cr := &configRoutes{}

	configGroup := router.Group("/config", auth.Middleware)
	configGroup.Get("/", wrap(cr.GetConfig))
	configGroup.Post("/update", wrap(cr.UpdateConfig))
}

var (
	val = validator.New()
)

func (cr *configRoutes) GetConfig(l *log.Logger, ctx *fiber.Ctx) error {
	cp := *config.I()
	cp.Secret = ""
	return ctx.JSON(cp)
}

func (cr *configRoutes) UpdateConfig(l *log.Logger, ctx *fiber.Ctx) error {
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		l.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	var c config.Config
	if err = ctx.BodyParser(&c); err != nil {
		l.Debug("Failed to parse config", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config"})
	}

	if err = val.Struct(c); err != nil {
		l.Debug("Invalid config", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err = config.I().Update(c, syncID); err != nil {
		l.Error("Failed to update config", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	log.Init(config.I().Logging)
	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}
