package routes

import (
	"errors"
	"path"
	"slices"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
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
	UnitOfWork          *db.UnitOfWork
}

func RegisterSubscriptionRoutes(sr subscriptionRoutes) {
	sr.Router.Group("/subscriptions", sr.Auth.Middleware).
		Get("/providers", sr.providers).
		Get("/all", withParam2(
			newQueryParam("allUsers", withAllowEmpty(false)),
			newQueryParam("", withAllowEmpty[utils.UserParams](), withStructConvertor[utils.UserParams]()),
			sr.all)).
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
	user := contextkey.GetFromContext(ctx, contextkey.User)
	allUsers = allUsers && user.HasRole(models.ManageSubscriptions)

	if allUsers {
		return sr.UnitOfWork.Subscriptions.All(ctx.UserContext())
	}

	return sr.UnitOfWork.Subscriptions.AllForUser(ctx.UserContext(), user.ID)
}

func (sr *subscriptionRoutes) providers(ctx *fiber.Ctx) error {
	return ctx.JSON(allowedProviders)
}

func (sr *subscriptionRoutes) runAll(ctx *fiber.Ctx, allUsers bool) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	subs, err := sr.getAll(ctx, allUsers)
	if err != nil {
		return InternalError(err)
	}

	for _, sub := range subs {
		err = sr.ContentService.DownloadSubscription(&sub, false) // This was manually triggered
		if err != nil {
			log.Error().Err(err).Msg("Failed to download subscription")
			return InternalError(err, fiber.Map{"subscription": sub.ID})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) runOnce(ctx *fiber.Ctx, id int) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)
	allowAny := user.HasRole(models.ManageSubscriptions)

	sub, err := sr.UnitOfWork.Subscriptions.Get(ctx.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return InternalError(err)
	}

	if sub.Owner != user.ID && !allowAny {
		return Forbidden()
	}

	err = sr.ContentService.DownloadSubscription(sub, false) // This was manually triggered
	if err != nil {
		log.Error().Err(err).Msg("Failed to download subscription")
		return InternalError(errors.New(sr.Transloco.GetTranslation("failed-to-run-once", err)))
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) all(ctx *fiber.Ctx, allUsers bool, userParams utils.UserParams) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)
	allUsers = allUsers && user.HasRole(models.ManageSubscriptions)

	var subs utils.PagedList[models.Subscription]
	var err error
	if allUsers {
		subs, err = sr.UnitOfWork.Subscriptions.AllPaginated(ctx.UserContext(), userParams)
	} else {
		subs, err = sr.UnitOfWork.Subscriptions.AllForUserPaginated(ctx.UserContext(), user.ID, userParams)
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscriptions")
		return InternalError(err)
	}

	return ctx.JSON(subs)
}

func (sr *subscriptionRoutes) get(ctx *fiber.Ctx, id int) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)
	allowAny := user.HasRole(models.ManageSubscriptions)

	sub, err := sr.UnitOfWork.Subscriptions.Get(ctx.UserContext(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return InternalError(err)
	}

	if sub.Owner != user.ID && !allowAny {
		return NotFound()
	}

	return ctx.JSON(sub)
}

func (sr *subscriptionRoutes) update(ctx *fiber.Ctx, sub models.Subscription) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)

	if err := sr.validatorSubscription(sub); err != nil {
		log.Error().Err(err).Msg("Failed to validate subscription")
		return BadRequest(err)
	}

	user := contextkey.GetFromContext(ctx, contextkey.User)
	allowAny := user.HasRole(models.ManageSubscriptions)

	cur, err := sr.UnitOfWork.Subscriptions.Get(ctx.UserContext(), sub.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscription")
		return InternalError(err)
	}

	if cur.Owner != user.ID && !allowAny {
		return Forbidden()
	}

	if err = sr.SubscriptionService.Update(ctx.UserContext(), sub); err != nil {
		log.Error().Err(err).Msg("Failed to update subscription")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(sub)
}

func (sr *subscriptionRoutes) new(ctx *fiber.Ctx, sub models.Subscription) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)
	sub.BaseDir = path.Clean(sub.BaseDir)

	if err := sr.validatorSubscription(sub); err != nil {
		log.Error().Err(err).Msg("Failed to validate subscription")
		return BadRequest(err)
	}

	// Force authenticated user
	sub.Owner = user.ID

	subscription, err := sr.SubscriptionService.Add(ctx.UserContext(), sub)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add subscription")
		return InternalError(err)
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

func (sr *subscriptionRoutes) delete(ctx *fiber.Ctx, id int) error {
	log := contextkey.GetFromContext(ctx, contextkey.Logger)
	user := contextkey.GetFromContext(ctx, contextkey.User)
	allowAny := user.HasRole(models.ManageSubscriptions)

	cur, err := sr.UnitOfWork.Subscriptions.Get(ctx.UserContext(), id)
	if err != nil {
		return InternalError(err)
	}

	if cur.Owner != user.ID && !allowAny {
		return Forbidden()
	}

	if err = sr.SubscriptionService.Delete(ctx.UserContext(), id); err != nil {
		log.Error().Err(err).Msg("Failed to delete subscription")
		return InternalError(err)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
