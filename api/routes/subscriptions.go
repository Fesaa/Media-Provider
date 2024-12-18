package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/subscriptions"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

type subscriptionRoutes struct {
	db *db.Database
}

func RegisterSubscriptionRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	sr := subscriptionRoutes{db: db}

	group := router.Group("/subscriptions", auth.Middleware)
	group.Get("/all", wrap(sr.All))
	group.Get("/:id", wrap(sr.Get))
	group.Post("/update", wrap(sr.Update))
	group.Post("/new", wrap(sr.New))
	group.Delete("/:id", wrap(sr.Delete))
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
	if err != nil {
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

	if err := sr.db.Subscriptions.Update(&sub); err != nil {
		l.Error("Failed to upsert subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	subscriptions.Refresh(sub.Id)

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

	subscription, err := sr.db.Subscriptions.New(sub)
	if err != nil {
		l.Error("Failed to create subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	subscriptions.Refresh(subscription.Id)
	return ctx.JSON(subscription)
}

func (sr *subscriptionRoutes) Delete(l *log.Logger, ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id", -1)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid id",
			"id":    utils.CopyString(ctx.Params("id", "")),
		})
	}

	if err = sr.db.Subscriptions.Delete(int64(id)); err != nil {
		l.Error("Failed to delete subscription", "error", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendStatus(fiber.StatusOK)
}
