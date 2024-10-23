package routes

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/gofiber/fiber/v2"
)

func (cr *contentRoutes) Search(l *log.Logger, ctx *fiber.Ctx) error {
	var searchRequest payload.SearchRequest
	if err := ctx.BodyParser(&searchRequest); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	search, err := providers.Search(searchRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(search)
}
