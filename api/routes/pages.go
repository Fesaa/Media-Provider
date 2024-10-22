package routes

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

func Pages(l *log.Logger, ctx *fiber.Ctx) error {
	pages, err := models.GetPages()
	if err != nil {
		l.Error("failed to retrieve pages", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(pages)
}

func Page(l *log.Logger, ctx *fiber.Ctx) error {
	id, _ := ctx.ParamsInt("index", -1)
	if id == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
		})
	}

	page, err := models.GetPage(int64(id))
	if err != nil {
		l.Error("failed to retrieve page", "error", err, slog.Int("pageId", id))
		return fiber.ErrInternalServerError
	}

	if page == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.JSON(page)
}
