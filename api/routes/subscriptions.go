package routes

import (
	"errors"
	"path"
	"slices"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

var (
	allowedProviders      = []models.Provider{models.MANGADEX, models.WEBTOON, models.DYNASTY, models.BATO}
	errDisallowedProvider = errors.New("the passed provider does not support subscription")
)

type subscriptionRoutes struct {
	dig.In

	Router fiber.Router
	Auth   services.AuthService

	Val                 services.ValidationService
	SubscriptionService services.SubscriptionService
	ContentService      services.ContentService
	Transloco           services.TranslocoService
}

func RegisterSubscriptionRoutes(sr subscriptionRoutes) {
	sr.Router.Group("/subscriptions", sr.Auth.Middleware).
		Get("/providers", sr.providers).
		Get("/all", withParam(newQueryParam("allUsers", withAllowEmpty(false)), sr.all)).
		Get("/:id", withParam(newIdPathParam(), sr.get)).
		Post("/run-once/:id", withParam(newIdPathParam(), sr.runOnce)).
		Post("/update", withBody(sr.update)).
		Post("/new", withBody(sr.new)).
		Post("/run-all", withParam(newQueryParam("allUsers", withAllowEmpty(false)), sr.runAll)).
		Delete("/:id", withParam(newIdPathParam(), sr.delete))
}

// getAll returns all subscriptions for the authenticated user. Or globally if allUsers ir true and the
// authenticated user as the ManageSubscriptions role
func (sr *subscriptionRoutes) getAll(ctx *fiber.Ctx, allUsers bool) ([]models.Subscription, error) {
	user := services.GetFromContext(ctx, services.UserKey)
	allUsers = allUsers && user.HasRole(models.ManageSubscriptions)

	if allUsers {
		return sr.SubscriptionService.All()
	}

	return sr.SubscriptionService.AllForUser(user.ID)
}

func (sr *subscriptionRoutes) providers(ctx *fiber.Ctx) error {
	return ctx.JSON(allowedProviders)
}

func (sr *subscriptionRoutes) runAll(ctx *fiber.Ctx, allUsers bool) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	subs, err := sr.getAll(ctx, allUsers)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	for _, sub := range subs {
		err = sr.ContentService.DownloadSubscription(&sub, false) // This was manually triggered
		if err != nil {
			log.Error().Err(err).Msg("Failed to download subscription")
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"sub-id":  sub.ID,
				"message": sr.Transloco.GetTranslation("failed-to-run-once", err),
			})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) runOnce(ctx *fiber.Ctx, id uint) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)
	allowAny := user.HasRole(models.ManageSubscriptions)

	sub, err := sr.SubscriptionService.Get(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if sub.Owner != user.ID && !allowAny {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{})
	}

	err = sr.ContentService.DownloadSubscription(sub, false) // This was manually triggered
	if err != nil {
		log.Error().Err(err).Msg("Failed to download subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": sr.Transloco.GetTranslation("failed-to-run-once", err),
		})
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) all(ctx *fiber.Ctx, allUsers bool) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	subs, err := sr.getAll(ctx, allUsers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscriptions")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(subs)
}

func (sr *subscriptionRoutes) get(ctx *fiber.Ctx, id uint) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)
	allowAny := user.HasRole(models.ManageSubscriptions)

	sub, err := sr.SubscriptionService.Get(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if sub.Owner != user.ID && !allowAny {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	return ctx.JSON(sub)
}

func (sr *subscriptionRoutes) update(ctx *fiber.Ctx, sub models.Subscription) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	if err := sr.validatorSubscription(sub); err != nil {
		log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	user := services.GetFromContext(ctx, services.UserKey)
	allowAny := user.HasRole(models.ManageSubscriptions)

	cur, err := sr.SubscriptionService.Get(sub.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if cur.Owner != user.ID && !allowAny {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{})
	}

	if err := sr.SubscriptionService.Update(sub); err != nil {
		log.Error().Err(err).Msg("Failed to update subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(sub)
}

func (sr *subscriptionRoutes) new(ctx *fiber.Ctx, sub models.Subscription) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)
	sub.Info.BaseDir = path.Clean(sub.Info.BaseDir)

	if err := sr.validatorSubscription(sub); err != nil {
		log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Force authenticated user
	sub.Owner = user.ID

	subscription, err := sr.SubscriptionService.Add(sub)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	go func() {
		if err = sr.ContentService.DownloadSubscription(subscription, false); err != nil {
			log.Warn().Err(err).Msg("failed to download subscription, will run again as scheduled. May have issues?")
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
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)
	allowAny := user.HasRole(models.ManageSubscriptions)

	cur, err := sr.SubscriptionService.Get(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	if cur.Owner != user.ID && !allowAny {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{})
	}

	if err := sr.SubscriptionService.Delete(id); err != nil {
		log.Error().Err(err).Msg("Failed to delete subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
