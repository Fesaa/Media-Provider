package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func Pages(ctx *fiber.Ctx) error {
	return ctx.JSON(config.Get(ctx).Pages)
}

func Page(ctx *fiber.Ctx) error {
	index, err := ctx.ParamsInt("index", -1)
	if err != nil || index == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid index",
		})
	}

	if index >= len(config.Get(ctx).Pages) || index < 0 {
		return ctx.Status(404).JSON(fiber.Map{
			"error": "Page not found",
		})
	}

	return ctx.JSON(config.Get(ctx).Pages[index])
}
