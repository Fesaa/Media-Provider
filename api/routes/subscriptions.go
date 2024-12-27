package routes

import (
	"errors"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/subscriptions"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"log/slog"
	"slices"
)

var (
	allowedProviders   = []models.Provider{models.MANGADEX, models.WEBTOON, models.DYNASTY}
	disallowedProvider = errors.New("the passed provider does not support subscription")
	notADir            = errors.New("the passed baseDir is not a directory")
)

type subscriptionRoutes struct {
	db *db.Database
}

func RegisterSubscriptionRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	sr := subscriptionRoutes{db: db}

	group := router.Group("/subscriptions", auth.Middleware)
	group.Get("/providers", sr.Providers)
	group.Get("/all", wrap(sr.All))
	group.Get("/:id", wrap(sr.Get))
	group.Post("/update", wrap(sr.Update))
	group.Post("/new", wrap(sr.New))
	group.Delete("/:id", wrap(sr.Delete))
	group.Post("/run-once/:id", wrap(sr.RunOnce))
}

func (sr *subscriptionRoutes) Providers(ctx *fiber.Ctx) error {
	return ctx.JSON(allowedProviders)
}

func (sr *subscriptionRoutes) RunOnce(l *log.Logger, ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id", -1)
	if err != nil || id == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
			"id":    utils.CopyString(ctx.Params("id", "")),
		})
	}

	sub, err := sr.db.Subscriptions.Get(int64(id))
	if err != nil {
		l.Error("Failed to get subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	err = providers.Download(payload.DownloadRequest{
		Id:        sub.ContentId,
		Provider:  sub.Provider,
		TempTitle: sub.Info.Title,
		BaseDir:   sub.Info.BaseDir,
	})
	if err != nil {
		l.Error("Failed to download subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{})
}

func (sr *subscriptionRoutes) All(l *log.Logger, ctx *fiber.Ctx) error {
	subs, err := sr.db.Subscriptions.All()
	if err != nil {
		l.Error("Failed to get subscriptions", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(subs)
}

func (sr *subscriptionRoutes) Get(l *log.Logger, ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id", -1)
	if err != nil || id == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
			"id":    utils.CopyString(ctx.Params("id", "")),
		})
	}

	sub, err := sr.db.Subscriptions.Get(int64(id))
	if err != nil {
		l.Error("Failed to get subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.JSON(sub)
}

func (sr *subscriptionRoutes) Update(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteConfig) {
		l.Warn("user does not have permission to edit subscriptions", "user", user.Name)
		return fiber.ErrUnauthorized
	}

	var sub models.Subscription
	if err := ctx.BodyParser(&sub); err != nil {
		l.Error("Failed to parse subscription body", "error", err)
		return fiber.ErrBadRequest
	}

	if err := sr.validatorSubscription(sub); err != nil {
		l.Error("Failed to validate subscription", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	prev, err := sr.db.Subscriptions.Get(sub.Id)
	if err != nil {
		l.Warn("Failed to get subscription", "error", err, slog.Int64("id", sub.Id))
		return fiber.ErrInternalServerError
	}

	if prev == nil {
		return fiber.ErrNotFound
	}

	if err = sr.db.Subscriptions.Update(sub); err != nil {
		l.Error("Failed to upsert subscription", "error", err, slog.Int64("id", sub.Id))
		return fiber.ErrInternalServerError
	}

	if sub.ShouldRefresh(prev) {
		subscriptions.Refresh(sub.Id)
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func (sr *subscriptionRoutes) New(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteConfig) {
		l.Warn("user does not have permission to create subscription", "user", user.Name)
		return fiber.ErrUnauthorized
	}

	var sub models.Subscription
	if err := ctx.BodyParser(&sub); err != nil {
		l.Error("Failed to parse subscription body", "error", err)
		return fiber.ErrBadRequest
	}

	sub.Info.BaseDir = CleanPath(sub.Info.BaseDir)

	if err := sr.validatorSubscription(sub); err != nil {
		l.Error("Failed to validate subscription", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	subscription, err := sr.db.Subscriptions.New(sub)
	if err != nil {
		l.Error("Failed to create subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	subscriptions.Refresh(subscription.Id)
	return ctx.JSON(subscription)
}

func (sr *subscriptionRoutes) validatorSubscription(sub models.Subscription) error {
	if err := val.Struct(&sub); err != nil {
		return err
	}

	if !slices.Contains(allowedProviders, sub.Provider) {
		return disallowedProvider
	}

	/*info, err := os.Stat(sub.Info.BaseDir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return notADir
	}*/

	return nil
}

func (sr *subscriptionRoutes) Delete(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteConfig) {
		l.Warn("user does not have permission to delete subscriptions", "user", user.Name)
		return fiber.ErrUnauthorized
	}

	id, err := ctx.ParamsInt("id", -1)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
			"id":    utils.CopyString(ctx.Params("id", "")),
		})
	}

	if id == -1 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
			"id":    utils.CopyString(ctx.Params("id", "")),
		})
	}

	if err = sr.db.Subscriptions.Delete(int64(id)); err != nil {
		l.Error("Failed to delete subscription", "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err = subscriptions.Delete(int64(id)); err != nil {
		l.Error("Failed to delete subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}
