package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"slices"
	"strings"
)

type pageRoutes struct {
	db *db.Database
}

func RegisterPageRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	pr := pageRoutes{db: db}

	pages := router.Group("/pages", auth.Middleware)
	pages.Get("/", wrap(pr.Pages))
	pages.Get("/:index", wrap(pr.Page))
	pages.Post("/new", wrap(pr.NewPage))
	pages.Post("/update", wrap(pr.UpdatePage))
	pages.Delete("/:pageId", wrap(pr.DeletePage))
	pages.Post("/swap", wrap(pr.SwapPage))
	pages.Post("/load-default", wrap(pr.LoadDefault))
}

func (pr *pageRoutes) Pages(l *log.Logger, ctx *fiber.Ctx) error {
	pages, err := pr.db.Pages.All()
	if err != nil {
		l.Error("failed to retrieve pages", "error", err)
		return fiber.ErrInternalServerError
	}

	slices.SortFunc(pages, func(a, b models.Page) int {
		sort := a.SortValue - b.SortValue
		if sort != 0 {
			return sort
		}

		return strings.Compare(a.Title, b.Title)
	})
	return ctx.JSON(pages)
}

func (pr *pageRoutes) Page(l *log.Logger, ctx *fiber.Ctx) error {
	id, _ := ctx.ParamsInt("index", -1)
	if id == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
		})
	}

	page, err := pr.db.Pages.Get(int64(id))
	if err != nil {
		l.Error("failed to retrieve page", "error", err, slog.Int("pageId", id))
		return fiber.ErrInternalServerError
	}

	if page == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.JSON(page)
}

func (pr *pageRoutes) UpdatePage(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWritePage) {
		l.Warn("user does not have permission to edit pages", "user", user.Name)
		return fiber.ErrUnauthorized
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

	if err := pr.db.Pages.Update(page); err != nil {
		l.Error("failed to insterts new page", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) NewPage(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWritePage) {
		l.Warn("user does not have permission to edit pages", "user", user.Name)
		return fiber.ErrUnauthorized
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

	if err := pr.db.Pages.New(page); err != nil {
		l.Error("failed to insert new page", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) DeletePage(l *log.Logger, ctx *fiber.Ctx) error {
	id, _ := ctx.ParamsInt("pageId", -1)
	if id == -1 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermDeletePage) {
		l.Warn("user does not have permission to delete page", "user", user.Name)
		return fiber.ErrUnauthorized
	}

	if err := pr.db.Pages.Delete(int64(id)); err != nil {
		l.Error("failed to delete page", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) SwapPage(l *log.Logger, ctx *fiber.Ctx) error {
	var m payload.SwapPageRequest
	if err := ctx.BodyParser(&m); err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	page1, err := pr.db.Pages.Get(m.Id1)
	if err != nil {
		l.Error("failed to retrieve page 1 by id", "error", err, slog.Int64("id1", m.Id1))
		return fiber.ErrInternalServerError
	}
	page2, err := pr.db.Pages.Get(m.Id2)
	if err != nil {
		l.Error("failed to retrieve page 2 by id", "error", err, slog.Int64("id2", m.Id2))
		return fiber.ErrInternalServerError
	}

	temp := page1.SortValue
	page1.SortValue = page2.SortValue
	page2.SortValue = temp

	if err = pr.db.Pages.Update(*page1); err != nil {
		l.Error("failed to update pages", "error", err)
		return fiber.ErrInternalServerError
	}
	if err = pr.db.Pages.Update(*page2); err != nil {
		l.Error("failed to upsert pages", "error", err)
		return fiber.ErrInternalServerError
	}
	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) LoadDefault(l *log.Logger, ctx *fiber.Ctx) error {
	pages, err := pr.db.Pages.All()
	if err != nil {
		l.Error("failed to retrieve pages", "error", err)
		return fiber.ErrInternalServerError
	}

	if len(pages) != 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot load default pages while other pages are present"})
	}

	for _, page := range models.DefaultPages {
		if err = pr.db.Pages.New(page); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Errorf("failed to load default pages %w", err).Error(),
			})
		}
	}

	return ctx.SendStatus(fiber.StatusOK)
}
