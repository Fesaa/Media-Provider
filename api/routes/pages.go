package routes

import (
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
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
	Auth   services.AuthService `name:"jwt-auth"`
	Log    zerolog.Logger

	Val            services.ValidationService
	PageService    services.PageService
	ContentService services.ContentService
	Transloco      services.TranslocoService
}

func RegisterPageRoutes(pr pageRoutes) {

	pages := pr.Router.Group("/pages", pr.Auth.Middleware)
	pages.Get("/", pr.Pages)
	pages.Post("/new", pr.UpdatePage)
	pages.Post("/update", pr.UpdatePage)
	pages.Delete("/:pageId", pr.DeletePage)
	pages.Post("/swap", pr.SwapPage)
	pages.Post("/load-default", pr.LoadDefault)
	pages.Get("/download-metadata", pr.DownloadMetadata)
	pages.Get("/:pageId", pr.Page)
}

func (pr *pageRoutes) Pages(ctx *fiber.Ctx) error {
	pages, err := pr.DB.Pages.All()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get pages")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
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
	id, err := ParamsUInt(ctx, "pageId")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	page, err := pr.DB.Pages.Get(id)
	if err != nil {
		pr.Log.Error().Err(err).Uint("pageId", id).Msg("Failed to get page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if page == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	return ctx.JSON(page)
}

func (pr *pageRoutes) UpdatePage(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWritePage) {
		pr.Log.Warn().Str("user", user.Name).Msg("user does not have page edit permission")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{})
	}

	var page models.Page
	if err := pr.Val.ValidateCtx(ctx, &page); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to parse page")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	page.Modifiers = utils.MapWithIdx(page.Modifiers, func(i int, mod models.Modifier) models.Modifier {
		mod.Sort = i

		// Ensure only one modifier value has the default state
		if mod.Type == models.DROPDOWN {
			mod.Values = utils.MapWithState(mod.Values, false, func(m models.ModifierValue, foundDefault bool) (models.ModifierValue, bool) {
				if foundDefault {
					m.Default = false
				}
				return m, foundDefault || m.Default
			})
		}

		return mod
	})

	if err := pr.PageService.UpdateOrCreate(&page); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to update page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(page)
}

func (pr *pageRoutes) DeletePage(ctx *fiber.Ctx) error {
	id, err := ParamsUInt(ctx, "pageId")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermDeletePage) {
		pr.Log.Warn().Str("user", user.Name).Msg("user does not have page delete permission")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{})
	}

	if err := pr.DB.Pages.Delete(id); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to delete page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) SwapPage(ctx *fiber.Ctx) error {
	var m payload.SwapPageRequest
	if err := pr.Val.ValidateCtx(ctx, &m); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to parse swap page")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := pr.PageService.SwapPages(m.Id1, m.Id2); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to swap page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) LoadDefault(ctx *fiber.Ctx) error {

	if err := pr.PageService.LoadDefaultPages(); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to load default pages")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) DownloadMetadata(ctx *fiber.Ctx) error {
	id := ctx.QueryInt("provider", -1)
	if id == -1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": pr.Transloco.GetTranslation("no-provider"),
		})
	}

	metadata, err := pr.ContentService.DownloadMetadata(models.Provider(id))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(metadata)
}
