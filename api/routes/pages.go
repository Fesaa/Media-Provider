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

func UpsertPage(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(*models.User)
	if !user.HasPermission(models.PermWritePage) {
		l.Warn("user does not have permission to edit pages", "user", user.Name)
		//return fiber.ErrUnauthorized
	}

	var page models.Page
	if err := ctx.BodyParser(&page); err != nil {
		l.Error("failed to parse request body", "error", err)
		return fiber.ErrBadRequest
	}

	if err := val.Struct(page); err != nil {
		log.Debug("page did not pass validation, contains errors", "error", err)
		return fiber.ErrBadRequest
	}

	if err := models.UpsertPage(&page); err != nil {
		l.Error("failed to upsert page", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func DeletePage(l *log.Logger, ctx *fiber.Ctx) error {
	id, _ := ctx.ParamsInt("pageId", -1)
	if id == -1 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermDeletePage) {
		l.Warn("user does not have permission to delete page", "user", user.Name)
		//return fiber.ErrUnauthorized
	}

	if err := models.DeletePageByID(int64(id)); err != nil {
		l.Error("failed to delete page", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}
