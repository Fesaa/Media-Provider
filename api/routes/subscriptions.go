package routes

import (
	"errors"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"path"
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

	sub, err := sr.SubscriptionService.Get(id)
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
	subs, err := sr.SubscriptionService.All()
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

	sub, err := sr.SubscriptionService.Get(id)
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

	if err := sr.SubscriptionService.Update(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to update subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
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

	sub.Info.BaseDir = path.Clean(sub.Info.BaseDir)

	if err := sr.validatorSubscription(sub); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to validate subscription")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	subscription, err := sr.SubscriptionService.Add(sub, true)
	if err != nil {
		sr.Log.Error().Err(err).Msg("Failed to add subscription")
		return fiber.ErrInternalServerError
	}

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

	if err = sr.SubscriptionService.Delete(id); err != nil {
		sr.Log.Error().Err(err).Msg("Failed to delete subscription")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
