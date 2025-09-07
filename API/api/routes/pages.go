package routes

import (
	"slices"
	"strings"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type pageRoutes struct {
	dig.In

	Router     fiber.Router
	UnitOfWork *db.UnitOfWork
	Auth       services.AuthService

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
		Post("/load-default", pr.loadDefault)

	pages.Use(hasRole(models.ManagePages)).
		Post("/new", withBodyValidation(pr.updatePage)).
		Post("/update", withBodyValidation(pr.updatePage)).
		Post("/order", withBody(pr.orderPages)).
		Delete("/:id", withParam(newIdPathParam(), pr.deletePage))
}

func (pr *pageRoutes) pages(ctx *fiber.Ctx) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)

	pages, err := pr.UnitOfWork.Pages.GetAllPages(ctx.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pages")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if len(user.Pages) > 0 {
		pages = utils.Filter(pages, func(page models.Page) bool {
			return slices.Contains(user.Pages, int32(page.ID))
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

func (pr *pageRoutes) page(ctx *fiber.Ctx, id int) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)

	if len(user.Pages) > 0 && !slices.Contains(user.Pages, int32(id)) {
		return NotFound()
	}

	page, err := pr.UnitOfWork.Pages.GetPage(ctx.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Int("pageId", id).Msg("Failed to get page")
		return InternalError(err)
	}

	if page == nil {
		return NotFound()
	}

	return ctx.JSON(page)
}

func (pr *pageRoutes) updatePage(ctx *fiber.Ctx, page models.Page) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

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

	cur, err := pr.UnitOfWork.Pages.GetPage(ctx.UserContext(), page.ID)
	if err != nil {
		log.Error().Err(err).Int("pageId", page.ID).Msg("Failed to get page")
		return InternalError(err)
	}

	// Do not allow changing sort value with update, should use /order
	if cur != nil {
		page.SortValue = cur.SortValue
	}

	if err = pr.PageService.UpdateOrCreate(ctx.UserContext(), &page); err != nil {
		log.Error().Err(err).Msg("Failed to update page")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(page)
}

func (pr *pageRoutes) deletePage(ctx *fiber.Ctx, id int) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	if err := pr.UnitOfWork.Pages.Delete(ctx.UserContext(), id); err != nil {
		log.Error().Err(err).Msg("Failed to delete page")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) orderPages(ctx *fiber.Ctx, order []int) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	if err := pr.PageService.OrderPages(ctx.UserContext(), order); err != nil {
		log.Error().Err(err).Msg("Failed to swap page")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) loadDefault(ctx *fiber.Ctx) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	if err := pr.PageService.LoadDefaultPages(ctx.UserContext()); err != nil {
		log.Error().Err(err).Msg("Failed to load default pages")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (pr *pageRoutes) DownloadMetadata(ctx *fiber.Ctx, provider int) error {
	metadata, err := pr.ContentService.DownloadMetadata(models.Provider(provider))
	if err != nil {
		return InternalError(err)
	}
	return ctx.Status(fiber.StatusOK).JSON(metadata)
}
