package services

import (
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const (
	SummarySize = 100
)

type Notifier interface {
	Notify(models.Notification)
}

type NotificationService interface {
	GetNotifications(models.User, time.Time) ([]models.Notification, error)

	Notify(models.Notification)

	// MarkRead marks the notification with id as read, and sends the NotificationRead event through SignalR
	MarkRead(models.User, uint) error
	// MarkReadMany marks all the notifications as read, and sends the NotificationRead event through SignalR
	MarkReadMany(models.User, []uint) error
	// MarkUnRead marks the notification with id as unread, and sends the Notification event through SignalR
	MarkUnRead(models.User, uint) error
	Delete(models.User, uint) error
	DeleteMany(models.User, []uint) error
}

func NotificationServiceProvider(log zerolog.Logger, db *db.Database, signalR SignalRService) NotificationService {
	return &notificationService{
		db:      db,
		log:     log.With().Str("handler", "notification-service").Logger(),
		signalR: signalR,
	}
}

type notificationService struct {
	db      *db.Database
	log     zerolog.Logger
	signalR SignalRService
}

func (n *notificationService) GetNotifications(user models.User, after time.Time) ([]models.Notification, error) {
	var notifications []models.Notification
	builder := n.db.DB().
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

func (n *notificationService) Notify(notification models.Notification) {
	n.log.Debug().Any("notification", notification).Msg("adding notification")
	n.signalR.Notify(notification)
	if err := n.db.Notifications.New(notification); err != nil {
		n.log.Error().Err(err).Msg("unable to add notification")
	}
}

func (n *notificationService) MarkRead(user models.User, id uint) error {
	notification, err := n.db.Notifications.Get(id)
	if err != nil {
		return err
	}

	if !notification.HasAccess(user) {
		return fiber.ErrUnauthorized
	}

	if err = n.db.Notifications.MarkRead(id); err != nil {
		return err
	}

	if notification.Group != models.GroupContent {
		n.signalR.Broadcast(payload.EvenTypeNotificationRead, fiber.Map{
			"amount": 1,
		})
	}

	return nil
}

func (n *notificationService) MarkReadMany(user models.User, ids []uint) error {
	notifications, err := n.db.Notifications.GetMany(ids)
	if err != nil {
		return err
	}

	if !utils.All(notifications, func(n models.Notification) bool {
		return n.HasAccess(user)
	}) {
		return fiber.ErrUnauthorized
	}

	if err = n.db.Notifications.MarkReadMany(ids); err != nil {
		return err
	}

	n.signalR.Broadcast(payload.EvenTypeNotificationRead, fiber.Map{
		"amount": utils.Count(notifications, func(notification models.Notification) bool {
			return notification.Group != models.GroupContent
		}),
	})

	return nil
}

func (n *notificationService) MarkUnRead(user models.User, id uint) error {
	notification, err := n.db.Notifications.Get(id)
	if err != nil {
		return err
	}

	if !notification.HasAccess(user) {
		return fiber.ErrUnauthorized
	}

	return n.db.Notifications.MarkUnread(id)
}

func (n *notificationService) Delete(user models.User, id uint) error {
	notification, err := n.db.Notifications.Get(id)
	if err != nil {
		return err
	}

	if !notification.HasAccess(user) {
		return fiber.ErrUnauthorized
	}

	return n.db.Notifications.Delete(id)
}

func (n *notificationService) DeleteMany(user models.User, ids []uint) error {
	notifications, err := n.db.Notifications.GetMany(ids)
	if err != nil {
		return err
	}

	if !utils.All(notifications, func(n models.Notification) bool {
		return n.HasAccess(user)
	}) {
		return fiber.ErrUnauthorized
	}

	return n.db.Notifications.DeleteMany(ids)
}
