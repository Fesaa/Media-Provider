package routes

import (
	"errors"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"slices"
)

var (
	allowedProviders      = []models.Provider{models.MANGADEX, models.WEBTOON, models.DYNASTY}
	errDisallowedProvider = errors.New("the passed provider does not support subscription")
)

type subscriptionRoutes struct {
	dig.In

	Router   fiber.Router
	Auth     auth.Provider `name:"jwt-auth"`
	DB       *db.Database
	Provider *providers.ContentProvider
	Log      zerolog.Logger
	Val      *validator.Validate

	SubscriptionService services.SubscriptionService
}

func RegisterSubscriptionRoutes(sr subscriptionRoutes) {
	group := sr.Router.Group("/subscriptions", sr.Auth.Middleware)
	group.Get("/providers", sr.Providers)
	group.Get("/all", sr.All)
	group.Get("/:id", sr.Get)
	group.Post("/update", sr.Update)
	group.Post("/new", sr.New)
	group.Delete("/:id", sr.Delete)
	group.Post("/run-once/:id", sr.RunOnce)
}

func (sr *subscriptionRoutes) Providers(ctx *fiber.Ctx) error {
	return ctx.JSON(allowedProviders)
}

func (sr *subscriptionRoutes) RunOnce(ctx *fiber.Ctx) error {
	id, err := ParamsUInt(ctx, "id")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "Invalid id",
			"id":      utils.CopyString(ctx.Params("id", "")),
		})
	}

	sub, err := sr.DB.Subscriptions.Get(id)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get subscription")
		return fiber.ErrInternalServerError
	}

	err = sr.Provider.Download(payload.DownloadRequest{
		Id:        sub.ContentId,
		Provider:  sub.Provider,
		TempTitle: sub.Info.Title,
		BaseDir:   sub.Info.BaseDir,
	})
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to download subscription")
		return fiber.ErrInternalServerError
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) All(ctx *fiber.Ctx) error {
	subs, err := sr.DB.Subscriptions.All()
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get subscriptions")
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(subs)
}

func (sr *subscriptionRoutes) Get(ctx *fiber.Ctx) error {
	id, err := ParamsUInt(ctx, "id")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "Invalid id",
			"id":      utils.CopyString(ctx.Params("id", "")),
		})
	}

	sub, err := sr.DB.Subscriptions.Get(id)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get subscription")
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(sub)
}

func (sr *subscriptionRoutes) Update(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteConfig) {
		sr.Log.Warn().Str("user", user.Name).Msg("user does not have permission to edit subscriptions")
		return fiber.ErrUnauthorized
	}

	var sub models.Subscription
	if err := ctx.BodyParser(&sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to parse subscription")
		return fiber.ErrBadRequest
	}

	if err := sr.validatorSubscription(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	prev, err := sr.DB.Subscriptions.Get(sub.ID)
	if err != nil {
		sr.Log.Error().Err(err).Uint("id", sub.ID).Msg("Failed to get subscription")
		return fiber.ErrInternalServerError
	}

	if prev == nil {
		return fiber.ErrNotFound
	}

	if err = sr.DB.Subscriptions.Update(sub); err != nil {
		sr.Log.Error().Err(err).Uint("id", sub.ID).Msg("Failed to update subscription")
		return fiber.ErrInternalServerError
	}

	if sub.ShouldRefresh(prev) {
		sr.SubscriptionService.Refresh(sub.ID, false)
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (sr *subscriptionRoutes) New(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteConfig) {
		sr.Log.Warn().Str("user", user.Name).Msg("user does not have permission to create subscriptions")
		return fiber.ErrUnauthorized
	}

	var sub models.Subscription
	if err := ctx.BodyParser(&sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to parse subscription")
		return fiber.ErrBadRequest
	}

	sub.Info.BaseDir = CleanPath(sub.Info.BaseDir)

	if err := sr.validatorSubscription(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	existing, err := sr.DB.Subscriptions.GetByContentId(sub.ContentId)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to get existing subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if existing != nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "Subscription for this contentID already exists",
		})
	}

	subscription, err := sr.DB.Subscriptions.New(sub)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to create subscription")
		return fiber.ErrInternalServerError
	}

	sr.SubscriptionService.Refresh(subscription.ID, true)
	return ctx.JSON(subscription)
}

func (sr *subscriptionRoutes) validatorSubscription(sub models.Subscription) error {
	if err := sr.Val.Struct(&sub); err != nil {
		return err
	}

	if !slices.Contains(allowedProviders, sub.Provider) {
		return errDisallowedProvider
	}

	return nil
}

func (sr *subscriptionRoutes) Delete(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteConfig) {
		sr.Log.Warn().Str("user", user.Name).Msg("user does not have permission to delete subscriptions")
		return fiber.ErrUnauthorized
	}

	id, err := ParamsUInt(ctx, "id")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"message": "Invalid id",
			"id":      utils.CopyString(ctx.Params("id", "")),
		})
	}

	if err = sr.DB.Subscriptions.Delete(id); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to delete subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err = sr.SubscriptionService.Delete(id); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to delete subscription")
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}
