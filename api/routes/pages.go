package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"slices"
	"strings"
)

type pageRoutes struct {
	dig.In

	Router fiber.Router
	DB     *db.Database
	Auth   auth.Provider `name:"jwt-auth"`
	Log    zerolog.Logger
	Val    *validator.Validate

	PageService services.PageService
}

func RegisterPageRoutes(pr pageRoutes) {

	pages := pr.Router.Group("/pages", pr.Auth.Middleware)
	pages.Get("/", pr.Pages)
	pages.Get("/:index", pr.Page)
	pages.Post("/new", pr.UpdatePage)
	pages.Post("/update", pr.UpdatePage)
	pages.Delete("/:pageId", pr.DeletePage)
	pages.Post("/swap", pr.SwapPage)
	pages.Post("/load-default", pr.LoadDefault)
}

func (pr *pageRoutes) Pages(ctx *fiber.Ctx) error {
	pages, err := pr.DB.Pages.All()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get pages")
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

func (pr *pageRoutes) Page(ctx *fiber.Ctx) error {
	id, _ := ctx.ParamsInt("index", -1)
	if id == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "Invalid id",
		})
	}

	page, err := pr.DB.Pages.Get(int64(id))
	if err != nil {
		pr.Log.Error().Err(err).Int("pageId", id).Msg("Failed to get page")
		return fiber.ErrInternalServerError
	}

	if page == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	return ctx.JSON(page)
}

func (pr *pageRoutes) UpdatePage(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWritePage) {
		pr.Log.Warn().Str("user", user.Name).Msg("user does not have page edit permission")
		return fiber.ErrUnauthorized
	}

	var page models.Page
	if err := ctx.BodyParser(&page); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to parse page")
		return fiber.ErrBadRequest
	}

	if err := pr.Val.Struct(page); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to validate page")
		return fiber.ErrBadRequest
	}

	if err := pr.DB.Pages.Update(page); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to update page")
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) DeletePage(ctx *fiber.Ctx) error {
	id, _ := ctx.ParamsInt("pageId", -1)
	if id == -1 {
		return ctx.SendStatus(fiber.StatusBadRequest)
	}

	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermDeletePage) {
		pr.Log.Warn().Str("user", user.Name).Msg("user does not have page delete permission")
		return fiber.ErrUnauthorized
	}

	if err := pr.DB.Pages.Delete(int64(id)); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to delete page")
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) SwapPage(ctx *fiber.Ctx) error {
	var m payload.SwapPageRequest
	if err := ctx.BodyParser(&m); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to parse swap page")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := pr.PageService.SwapPages(m.Id1, m.Id2); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to swap page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to swap page",
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) LoadDefault(ctx *fiber.Ctx) error {

	if err := pr.PageService.LoadDefaultPages(); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to load default pages")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load default pages",
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
