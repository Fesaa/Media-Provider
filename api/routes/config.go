package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

var (
	val = validator.New()
)

func GetConfig(ctx *fiber.Ctx) error {
	cp := *config.I()
	cp.Secret = ""
	return ctx.JSON(cp)
}

func UpdateConfig(ctx *fiber.Ctx) error {
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	var c config.Config
	if err = ctx.BodyParser(&c); err != nil {
		log.Debug("Failed to parse config", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config"})
	}

	if err = val.Struct(c); err != nil {
		log.Debug("Invalid config", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err = config.I().Update(c, syncID); err != nil {
		log.Error("Failed to update config", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	log.Init(config.I().Logging)
	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}
