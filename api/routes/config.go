package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

var (
	val = validator.New()
)

func GetConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(config.I())
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

	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}

func RemovePage(ctx *fiber.Ctx) error {
	index, err := intParam(ctx, "index")
	if err != nil {
		log.Debug("Invalid index", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid index"})
	}
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	if err = config.I().RemovePage(index, syncID); err != nil {
		log.Error("Failed to update page", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}

func AddPage(ctx *fiber.Ctx) error {
	var page config.Page
	err := ctx.BodyParser(&page)
	if err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	if err = val.Struct(page); err != nil {
		log.Debug("Invalid page", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err = config.I().AddPage(page, syncID); err != nil {
		log.Error("Failed to update page", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}

func UpdatePage(ctx *fiber.Ctx) error {
	var page config.Page
	err := ctx.BodyParser(&page)
	if err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	index, err := intParam(ctx, "index")
	if err != nil {
		log.Debug("Invalid index", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid index"})
	}
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	if err = val.Struct(page); err != nil {
		log.Debug("Invalid page", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err = config.I().UpdatePage(page, index, syncID); err != nil {
		log.Error("Failed to update page", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}

func MovePage(ctx *fiber.Ctx) error {
	var m payload.MovePageRequest
	if err := ctx.BodyParser(&m); err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	syncID, err := intQuery(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}
	if err = config.I().MovePage(m.OldIndex, m.NewIndex, syncID); err != nil {
		log.Error("Failed to save config", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).SendString(strconv.Itoa(config.I().SyncId))
}
