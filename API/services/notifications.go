package services

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type Notifier interface {
	Notify(context.Context, models.Notification)
}

type NotificationService interface {
	GetNotifications(context.Context, models.User, time.Time) ([]models.Notification, error)

	Notify(context.Context, models.Notification)

	// MarkRead marks the notification with id as read, and sends the NotificationRead event through SignalR
	MarkRead(context.Context, models.User, int) error
	// MarkReadMany marks all the notifications as read, and sends the NotificationRead event through SignalR
	MarkReadMany(context.Context, models.User, []int) error
	// MarkUnRead marks the notification with id as unread, and sends the Notification event through SignalR
	MarkUnRead(context.Context, models.User, int) error
	Delete(context.Context, models.User, int) error
	DeleteMany(context.Context, models.User, []int) error
}

func NotificationServiceProvider(log zerolog.Logger, unitOfWork *db.UnitOfWork, signalR SignalRService) NotificationService {
	return &notificationService{
		unitOfWork: unitOfWork,
		log:        log.With().Str("handler", "notification-service").Logger(),
		signalR:    signalR,
	}
}

type notificationService struct {
	unitOfWork *db.UnitOfWork
	log        zerolog.Logger
	signalR    SignalRService
}

func (n *notificationService) GetNotifications(ctx context.Context, user models.User, after time.Time) ([]models.Notification, error) {
	var notifications []models.Notification
	builder := n.unitOfWork.DB().
		WithContext(ctx).
		Where("owner IS NULL").
		Or("owner = ?", user.ID)

	if !after.IsZero() {
		builder = builder.Where("created_at > ?", after)
	}

	res := builder.Find(&notifications)

	if res.Error != nil {
		return nil, res.Error
	}

	roles := utils.MapToString(user.Roles)
	notifications = utils.Filter(notifications, func(n models.Notification) bool {
		if len(n.RequiredRoles) == 0 {
			return true
		}

		return utils.Contains(n.RequiredRoles, roles)
	})

	return notifications, nil
}

func (n *notificationService) Notify(ctx context.Context, notification models.Notification) {
	n.log.Debug().Any("notification", notification).Msg("adding notification")
	n.signalR.Notify(ctx, notification)
	if err := n.unitOfWork.Notifications.New(ctx, notification); err != nil {
		n.log.Error().Err(err).Msg("unable to add notification")
	}
}

func (n *notificationService) MarkRead(ctx context.Context, user models.User, id int) error {
	notification, err := n.unitOfWork.Notifications.Get(ctx, id)
	if err != nil {
		return err
	}

	if !notification.HasAccess(user) {
		return fiber.ErrUnauthorized
	}

	if err = n.unitOfWork.Notifications.MarkRead(ctx, id); err != nil {
		return err
	}

	if notification.Group != models.GroupContent {
		n.signalR.Broadcast(payload.EvenTypeNotificationRead, fiber.Map{
			"amount": 1,
		})
	}

	return nil
}

func (n *notificationService) MarkReadMany(ctx context.Context, user models.User, ids []int) error {
	notifications, err := n.unitOfWork.Notifications.GetMany(ctx, ids)
	if err != nil {
		return err
	}

	if !utils.All(notifications, func(n models.Notification) bool {
		return n.HasAccess(user)
	}) {
		return fiber.ErrUnauthorized
	}

	if err = n.unitOfWork.Notifications.MarkReadMany(ctx, ids); err != nil {
		return err
	}

	n.signalR.Broadcast(payload.EvenTypeNotificationRead, fiber.Map{
		"amount": utils.Count(notifications, func(notification models.Notification) bool {
			return notification.Group != models.GroupContent
		}),
	})

	return nil
}

func (n *notificationService) MarkUnRead(ctx context.Context, user models.User, id int) error {
	notification, err := n.unitOfWork.Notifications.Get(ctx, id)
	if err != nil {
		return err
	}

	if !notification.HasAccess(user) {
		return fiber.ErrUnauthorized
	}

	return n.unitOfWork.Notifications.MarkUnread(ctx, id)
}

func (n *notificationService) Delete(ctx context.Context, user models.User, id int) error {
	notification, err := n.unitOfWork.Notifications.Get(ctx, id)
	if err != nil {
		return err
	}

	if !notification.HasAccess(user) {
		return fiber.ErrUnauthorized
	}

	return n.unitOfWork.Notifications.Delete(ctx, id)
}

func (n *notificationService) DeleteMany(ctx context.Context, user models.User, ids []int) error {
	notifications, err := n.unitOfWork.Notifications.GetMany(ctx, ids)
	if err != nil {
		return err
	}

	if !utils.All(notifications, func(n models.Notification) bool {
		return n.HasAccess(user)
	}) {
		return fiber.ErrUnauthorized
	}

	return n.unitOfWork.Notifications.DeleteMany(ctx, ids)
}
