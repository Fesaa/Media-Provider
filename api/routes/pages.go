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

	userName := ctx.Locals("user").(string)
	user, err := models.GetUser(userName)
	if err != nil {
		l.Error("failed to retrieve user", "error", err)
		return fiber.ErrInternalServerError
	}

	if !user.HasPermission(models.PermDeletePage) {
		l.Warn("user does not have permission to delete page", "user", userName)
		//return fiber.ErrForbidden
	}

	err = models.DeletePageByID(int64(id))
	if err != nil {
		l.Error("failed to delete page", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}
