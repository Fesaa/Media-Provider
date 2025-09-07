package routes

import (
	"errors"
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type notificationRoutes struct {
	dig.In

	UnitOfWork *db.UnitOfWork
	Router     fiber.Router
	Auth       services.AuthService

	NotificationService services.NotificationService
	Transloco           services.TranslocoService
}

func RegisterNotificationRoutes(nr notificationRoutes) {
	nr.Router.Group("/notifications", nr.Auth.Middleware).
		Get("/all", withParam(newQueryParam("after", withAllowEmpty(time.Time{})), nr.all)).
		Get("/recent", withParam(newQueryParam("limit", withAllowEmpty(5)), nr.recent)).
		Get("/amount", nr.amount).
		Post("/:id/read", withParam(newIdPathParam(), nr.read)).
		Post("/:id/unread", withParam(newIdPathParam(), nr.unread)).
		Delete("/:id", withParam(newIdPathParam(), nr.delete)).
		Post("/many", withBody(nr.readMany)).
		Post("/many/delete", withBody(nr.deleteMany))
}

func (nr *notificationRoutes) all(ctx *fiber.Ctx, after time.Time) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)
	notifications, err := nr.NotificationService.GetNotifications(ctx.UserContext(), user, after)

	if err != nil {
		log.Error().Err(err).Time("after", after).Msg("failed to fetch notifications")
		return InternalError(err)
	}

	return ctx.JSON(notifications)
}

func (nr *notificationRoutes) recent(ctx *fiber.Ctx, limit int) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	amount := utils.Clamp(limit, 1, 10)
	nots, err := nr.UnitOfWork.Notifications.Recent(ctx.UserContext(), amount, models.GroupContent)
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch recent notifications")
		return InternalError(err)
	}

	return ctx.JSON(nots)
}

func (nr *notificationRoutes) amount(ctx *fiber.Ctx) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	size, err := nr.UnitOfWork.Notifications.Unread(ctx.UserContext())
	if err != nil {
		log.Error().Err(err).Msg("failed to fetch amount of unread notifications")
		return InternalError(err)
	}
	return ctx.JSON(size)
}

func (nr *notificationRoutes) read(ctx *fiber.Ctx, id int) error {
	user := services.GetFromContext(ctx, services.UserKey)
	err := nr.NotificationService.MarkRead(ctx.UserContext(), user, id)
	return nr.handleServiceError(ctx, err)
}

func (nr *notificationRoutes) readMany(ctx *fiber.Ctx, ids []int) error {
	user := services.GetFromContext(ctx, services.UserKey)
	err := nr.NotificationService.MarkReadMany(ctx.UserContext(), user, ids)
	return nr.handleServiceError(ctx, err)
}

func (nr *notificationRoutes) unread(ctx *fiber.Ctx, id int) error {
	user := services.GetFromContext(ctx, services.UserKey)
	err := nr.NotificationService.MarkUnRead(ctx.UserContext(), user, id)
	return nr.handleServiceError(ctx, err)
}

func (nr *notificationRoutes) deleteMany(ctx *fiber.Ctx, ids []int) error {
	user := services.GetFromContext(ctx, services.UserKey)
	err := nr.NotificationService.DeleteMany(ctx.UserContext(), user, ids)
	return nr.handleServiceError(ctx, err)
}

func (nr *notificationRoutes) delete(ctx *fiber.Ctx, id int) error {
	user := services.GetFromContext(ctx, services.UserKey)
	err := nr.NotificationService.Delete(ctx.UserContext(), user, id)
	return nr.handleServiceError(ctx, err)
}

func (nr *notificationRoutes) handleServiceError(ctx *fiber.Ctx, err error) error {
	if err == nil {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
	}

	if errors.Is(err, fiber.ErrUnauthorized) {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{})
	}

	return InternalError(err)
}
