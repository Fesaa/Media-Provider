package routes

import (
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/gofiber/fiber/v2"
)

func Search(ctx *fiber.Ctx) error {
	var searchRequest providers.SearchRequest
	if err := ctx.BodyParser(&searchRequest); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	search, err := providers.Search(searchRequest)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{})
	}
	return ctx.JSON(search)
}
