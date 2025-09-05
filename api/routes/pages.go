package routes

import (
	"slices"
	"strings"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type pageRoutes struct {
	dig.In

	Router fiber.Router
	DB     *db.Database
	Auth   services.AuthService
	Log    zerolog.Logger

	Val            services.ValidationService
	PageService    services.PageService
	ContentService services.ContentService
	Transloco      services.TranslocoService
}

func RegisterPageRoutes(pr pageRoutes) {
	pages := pr.Router.Group("/pages", pr.Auth.Middleware)
	pages.
		Get("/", pr.pages).
		Get("/download-metadata", withParam(newQueryParam("provider",
			withMessage[int](pr.Transloco.GetTranslation("no-provider"))), pr.DownloadMetadata)).
		Get("/:id", withParam(newIdPathParam(), pr.page)).
		Post("/order", withBody(pr.orderPages)).
		Post("/load-default", pr.loadDefault)

	pages.Use(hasRole(models.ManagePages)).
		Post("/new", withBodyValidation(pr.updatePage)).
		Post("/update", withBodyValidation(pr.updatePage)).
		Delete("/:id", withParam(newIdPathParam(), pr.deletePage))
}

func (pr *pageRoutes) pages(ctx *fiber.Ctx) error {
	user := services.GetFromContext(ctx, services.UserKey)

	pages, err := pr.DB.Pages.All()
	if err != nil {
		pr.Log.Error().Err(err).Msg("Failed to get pages")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if len(user.Pages) > 0 {
		pages = utils.Filter(pages, func(page models.Page) bool {
			return slices.Contains(user.Pages, int32(page.ID)) //nolint:gosec
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

func (pr *pageRoutes) page(ctx *fiber.Ctx, id uint) error {
	user := services.GetFromContext(ctx, services.UserKey)

	if len(user.Pages) > 0 && !slices.Contains(user.Pages, int32(id)) { //nolint:gosec
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
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

func (pr *pageRoutes) updatePage(ctx *fiber.Ctx, page models.Page) error {
	page.Modifiers = utils.MapWithIdx(page.Modifiers, func(i int, mod models.Modifier) models.Modifier {
		mod.Sort = i

		// Ensure only one modifier value has the default state
		if mod.Type == models.DROPDOWN {
			mod.Values = utils.MapWithState(mod.Values, false,
				func(m models.ModifierValue, foundDefault bool) (models.ModifierValue, bool) {
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

func (pr *pageRoutes) deletePage(ctx *fiber.Ctx, id uint) error {
	if err := pr.DB.Pages.Delete(id); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to delete page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) orderPages(ctx *fiber.Ctx, order []uint) error {
	if err := pr.PageService.OrderPages(order); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to swap page")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) loadDefault(ctx *fiber.Ctx) error {
	if err := pr.PageService.LoadDefaultPages(); err != nil {
		pr.Log.Error().Err(err).Msg("Failed to load default pages")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) DownloadMetadata(ctx *fiber.Ctx, provider int) error {
	metadata, err := pr.ContentService.DownloadMetadata(models.Provider(provider))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(metadata)
}
