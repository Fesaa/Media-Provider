package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
)

func GetConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(config.I())
}

func RemovePage(ctx *fiber.Ctx) error {
	index, err := intParam(ctx, "index")
	if err != nil {
		log.Debug("Invalid index", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid index"})
	}
	syncID, err := intParam(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	if err = config.I().RemovePage(index, syncID); err != nil {
		log.Error("Failed to update page", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func AddPage(ctx *fiber.Ctx) error {
	var page config.Page
	err := ctx.BodyParser(&page)
	if err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	syncID, err := intParam(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	if err = config.I().AddPage(page, syncID); err != nil {
		log.Error("Failed to update page", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
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
	syncID, err := intParam(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}

	if err = config.I().UpdatePage(page, index, syncID); err != nil {
		log.Error("Failed to update page", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func MovePage(ctx *fiber.Ctx) error {
	oldIndex, err := intParam(ctx, "old_index")
	if err != nil {
		log.Debug("Invalid index", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid index"})
	}
	newIndex, err := intParam(ctx, "old_index")
	if err != nil {
		log.Debug("Invalid index", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid index"})
	}
	syncID, err := intParam(ctx, "sync_id")
	if err != nil {
		log.Debug("Invalid sync_id", "error", err)
		return ctx.Status(fiber.StatusPreconditionRequired).JSON(fiber.Map{"error": "Invalid sync_id"})
	}
	if err = config.I().MovePage(oldIndex, newIndex, syncID); err != nil {
		log.Error("Failed to save config", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
