package routes

import (
	"time"

	"github.com/Fesaa/Media-Provider/api/middleware"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type notificationRoutes struct {
	dig.In

	DB     *db.Database
	Router fiber.Router
	Auth   services.AuthService `name:"jwt-auth"`
	Log    zerolog.Logger

	NotificationService services.NotificationService
	Transloco           services.TranslocoService
}

func RegisterNotificationRoutes(nr notificationRoutes) {
	notificationGroup := nr.Router.Group("/notifications", nr.Auth.Middleware)
	notificationGroup.Get("/all", middleware.WithQueryParams(middleware.WithAllowEmpty[time.Time]("after"), nr.all))
	notificationGroup.Get("/recent", middleware.WithQueryParams(middleware.WithAllowEmpty("limit", 5), nr.recent))
	notificationGroup.Get("/amount", nr.amount)
	notificationGroup.Post("/:id/read", middleware.WithParams(middleware.WithQueryName[uint]("id"), nr.read))
	notificationGroup.Post("/:id/unread", middleware.WithParams(middleware.WithQueryName[uint]("id"), nr.unread))
	notificationGroup.Delete("/:id", middleware.WithParams(middleware.WithQueryName[uint]("id"), nr.delete))
	notificationGroup.Post("/many", middleware.WithBody(nr.readMany))
	notificationGroup.Post("/many/delete", middleware.WithBody(nr.deleteMany))
}

func (nr *notificationRoutes) all(ctx *fiber.Ctx, after time.Time) error {
	var notifications []models.Notification
	var err error

	if after.IsZero() {
		notifications, err = nr.DB.Notifications.All()
	} else {
		notifications, err = nr.DB.Notifications.AllAfter(after)
	}

	if err != nil {
		nr.Log.Error().Err(err).Time("after", after).Msg("failed to fetch notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(notifications)
}

func (nr *notificationRoutes) recent(ctx *fiber.Ctx, limit int) error {
	nots, err := nr.DB.Notifications.Recent(utils.Clamp(limit, 1, 10), models.GroupContent)
	if err != nil {
		nr.Log.Error().Err(err).Msg("failed to fetch recent notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(nots)
}

func (nr *notificationRoutes) amount(ctx *fiber.Ctx) error {
	size, err := nr.DB.Notifications.Unread()
	if err != nil {
		nr.Log.Error().Err(err).Msg("failed to fetch amount of unread notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.JSON(size)
}

func (nr *notificationRoutes) read(ctx *fiber.Ctx, id uint) error {
	if err := nr.NotificationService.MarkRead(id); err != nil {
		nr.Log.Error().Err(err).Msg("failed to mark notification read")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) readMany(ctx *fiber.Ctx, ids []uint) error {
	if err := nr.NotificationService.MarkReadMany(ids); err != nil {
		nr.Log.Error().Err(err).Msg("failed to mark notifications read")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) unread(ctx *fiber.Ctx, id uint) error {
	if err := nr.NotificationService.MarkUnRead(id); err != nil {
		nr.Log.Error().Err(err).Msg("failed to mark notification unread")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) deleteMany(ctx *fiber.Ctx, ids []uint) error {
	if err := nr.DB.Notifications.DeleteMany(ids); err != nil {
		nr.Log.Error().Err(err).Msg("failed to delete notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) delete(ctx *fiber.Ctx, id uint) error {
	if err := nr.DB.Notifications.Delete(id); err != nil {
		nr.Log.Error().Err(err).Msg("failed to delete notification")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
