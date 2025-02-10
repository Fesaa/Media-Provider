package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"time"
)

type notificationRoutes struct {
	dig.In

	DB     *db.Database
	Router fiber.Router
	Auth   auth.Provider `name:"jwt-auth"`
	Log    zerolog.Logger

	NotificationService services.NotificationService
}

func RegisterNotificationRoutes(nr notificationRoutes) {
	notificationGroup := nr.Router.Group("/notifications", nr.Auth.Middleware)
	notificationGroup.Get("/all", nr.All)
	notificationGroup.Get("/amount", nr.Amount)
	notificationGroup.Post("/:id/read", nr.Read)
	notificationGroup.Post("/:id/unread", nr.Unread)
	notificationGroup.Delete("/:id", nr.Delete)
	notificationGroup.Post("/many", nr.ReadMany)
	notificationGroup.Post("/many/delete", nr.DeleteMany)
}

func (nr *notificationRoutes) All(ctx *fiber.Ctx) error {
	var notifications []models.Notification
	var err error

	timeS := ctx.Query("after", "")
	if timeS != "" {
		var t time.Time
		t, err = time.Parse(time.RFC3339, timeS)
		if err != nil {
			nr.Log.Error().Err(err).Str("after", timeS).Msg("failed to parse passed time")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid time format: " + err.Error(),
			})
		}

		notifications, err = nr.DB.Notifications.AllAfter(t)
	} else {
		notifications, err = nr.DB.Notifications.All()
	}

	if err != nil {
		nr.Log.Error().Err(err).Str("after", timeS).Msg("failed to fetch notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.JSON(notifications)
}

func (nr *notificationRoutes) Amount(ctx *fiber.Ctx) error {
	size, err := nr.DB.Notifications.Unread()
	if err != nil {
		nr.Log.Error().Err(err).Msg("failed to fetch amount of unread notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.JSON(size)
}

func (nr *notificationRoutes) Read(ctx *fiber.Ctx) error {
	id, err := ParamsUInt(ctx, "id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	err = nr.NotificationService.MarkRead(id)
	if err != nil {
		nr.Log.Error().Err(err).Msg("failed to mark notification read")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) ReadMany(ctx *fiber.Ctx) error {
	var ids []uint
	if err := ctx.BodyParser(&ids); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := nr.DB.Notifications.MarkReadMany(ids); err != nil {
		nr.Log.Error().Err(err).Msg("failed to mark notifications read")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) Unread(ctx *fiber.Ctx) error {
	id, err := ParamsUInt(ctx, "id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	err = nr.NotificationService.MarkUnRead(id)
	if err != nil {
		nr.Log.Error().Err(err).Msg("failed to mark notification unread")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) DeleteMany(ctx *fiber.Ctx) error {
	var ids []uint
	if err := ctx.BodyParser(&ids); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := nr.DB.Notifications.DeleteMany(ids); err != nil {
		nr.Log.Error().Err(err).Msg("failed to delete notifications")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (nr *notificationRoutes) Delete(ctx *fiber.Ctx) error {
	id, err := ParamsUInt(ctx, "id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	err = nr.DB.Notifications.Delete(id)
	if err != nil {
		nr.Log.Error().Err(err).Msg("failed to delete notification")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
