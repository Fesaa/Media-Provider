package routes

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
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
}

func RegisterPageRoutes(pr pageRoutes) {

	pages := pr.Router.Group("/pages", pr.Auth.Middleware)
	pages.Get("/", pr.Pages)
	pages.Get("/:index", pr.Page)
	pages.Post("/new", pr.NewPage)
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
			"error": "Invalid id",
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

func (pr *pageRoutes) NewPage(ctx *fiber.Ctx) error {
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

	if err := pr.DB.Pages.New(page); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to create page")
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
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	page1, err := pr.DB.Pages.Get(m.Id1)
	if err != nil {
		pr.Log.Error().Err(err).Int64("id", m.Id1).Msg("Failed to get page1")
		return fiber.ErrInternalServerError
	}
	page2, err := pr.DB.Pages.Get(m.Id2)
	if err != nil {
		pr.Log.Error().Err(err).Int64("id", m.Id2).Msg("Failed to get page2")
		return fiber.ErrInternalServerError
	}

	temp := page1.SortValue
	page1.SortValue = page2.SortValue
	page2.SortValue = temp

	if err = pr.DB.Pages.Update(*page1); err != nil {
		pr.Log.Error().Err(err).Int64("id", m.Id1).Msg("Failed to update page1")
		return fiber.ErrInternalServerError
	}
	if err = pr.DB.Pages.Update(*page2); err != nil {
		pr.Log.Error().Err(err).Int64("id", m.Id2).Msg("Failed to update page2")
		return fiber.ErrInternalServerError
	}
	return ctx.SendStatus(fiber.StatusOK)
}

func (pr *pageRoutes) LoadDefault(ctx *fiber.Ctx) error {
	pages, err := pr.DB.Pages.All()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get pages")
		return fiber.ErrInternalServerError
	}

	if len(pages) != 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot load default pages while other pages are present"})
	}

	for _, page := range models.DefaultPages {
		if err = pr.DB.Pages.New(page); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Errorf("failed to load default pages %w", err).Error(),
			})
		}
	}

	return ctx.SendStatus(fiber.StatusOK)
}
