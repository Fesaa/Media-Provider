package routes

import (
	"errors"
	"path"
	"slices"

	"github.com/Fesaa/Media-Provider/api/middleware"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

var (
	allowedProviders      = []models.Provider{models.MANGADEX, models.WEBTOON, models.DYNASTY, models.BATO}
	errDisallowedProvider = errors.New("the passed provider does not support subscription")
)

type subscriptionRoutes struct {
	dig.In

	Router fiber.Router
	Auth   services.AuthService `name:"jwt-auth"`
	Log    zerolog.Logger

	Val                 services.ValidationService
	SubscriptionService services.SubscriptionService
	ContentService      services.ContentService
	Transloco           services.TranslocoService
}

func RegisterSubscriptionRoutes(sr subscriptionRoutes) {
	group := sr.Router.Group("/subscriptions", sr.Auth.Middleware)
	group.Get("/providers", sr.providers)
	group.Get("/all", sr.all)
	group.Get("/:id", middleware.WithParams(middleware.IdParamsOption(), sr.get))

	group.Use(middleware.HasRole(models.ManageSubscriptions)).
		Post("/update", middleware.WithBody(sr.update)).
		Post("/new", middleware.WithBody(sr.new)).
		Delete("/:id", middleware.WithParams(middleware.IdParamsOption(), sr.delete)).
		Post("/run-once/:id", middleware.WithParams(middleware.IdParamsOption(), sr.runOnce)).
		Post("/run-all", sr.runAll)
}

func (sr *subscriptionRoutes) providers(ctx *fiber.Ctx) error {
	return ctx.JSON(allowedProviders)
}

func (sr *subscriptionRoutes) runAll(ctx *fiber.Ctx) error {
	subs, err := sr.SubscriptionService.All()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	for _, sub := range subs {
		err = sr.ContentService.DownloadSubscription(&sub, false) // This was manually triggered
		if err != nil {
			sr.Log.Error().Err(err).Msg("Failed to download subscription")
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"sub-id":  sub.ID,
				"message": sr.Transloco.GetTranslation("failed-to-run-once", err),
			})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) runOnce(ctx *fiber.Ctx, id uint) error {
	sub, err := sr.SubscriptionService.Get(id)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	err = sr.ContentService.DownloadSubscription(sub, false) // This was manually triggered
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to download subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": sr.Transloco.GetTranslation("failed-to-run-once", err),
		})
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) all(ctx *fiber.Ctx) error {
	subs, err := sr.SubscriptionService.All()
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get subscriptions")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(subs)
}

func (sr *subscriptionRoutes) get(ctx *fiber.Ctx, id uint) error {
	sub, err := sr.SubscriptionService.Get(id)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(sub)
}

func (sr *subscriptionRoutes) update(ctx *fiber.Ctx, sub models.Subscription) error {
	if err := sr.validatorSubscription(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := sr.SubscriptionService.Update(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to update subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(sub)
}

func (sr *subscriptionRoutes) new(ctx *fiber.Ctx, sub models.Subscription) error {
	sub.Info.BaseDir = path.Clean(sub.Info.BaseDir)

	if err := sr.validatorSubscription(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	subscription, err := sr.SubscriptionService.Add(sub)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to add subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	go func() {
		if err = sr.ContentService.DownloadSubscription(subscription, false); err != nil {
			sr.Log.Warn().Err(err).Msg("failed to download subscription, will run again as scheduled. May have issues?")
		}
	}()

	return ctx.JSON(subscription)
}

func (sr *subscriptionRoutes) validatorSubscription(sub models.Subscription) error {
	if err := sr.Val.Validate(sub); err != nil {
		return err
	}

	if !slices.Contains(allowedProviders, sub.Provider) {
		return errDisallowedProvider
	}

	return nil
}

func (sr *subscriptionRoutes) delete(ctx *fiber.Ctx, id uint) error {
	if err := sr.SubscriptionService.Delete(id); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to delete subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
