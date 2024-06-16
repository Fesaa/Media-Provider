package routes

import (
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/gofiber/fiber/v2"
)

func Search(ctx *fiber.Ctx) error {
	var searchRequest payload.SearchRequest
	if err := ctx.BodyParser(&searchRequest); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	_, err := providers.Search(searchRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "this pls",
	})
}
