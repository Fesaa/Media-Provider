package services

import (
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
	Notify(models.Notification)

	// MarkRead marks the notification with id as read, and sends the NotificationRead event through SignalR
	MarkRead(id uint) error
	// MarkReadMany marks all the notifications as read, and sends the NotificationRead event through SignalR
	MarkReadMany([]uint) error
	// MarkUnRead marks the notification with id as unread, and sends the Notification event through SignalR
	MarkUnRead(id uint) error
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

func (n *notificationService) Notify(notification models.Notification) {
	n.log.Debug().Any("notification", notification).Msg("adding notification")
	n.signalR.Notify(notification)
	if err := n.db.Notifications.New(notification); err != nil {
		n.log.Error().Err(err).Msg("unable to add notification")
	}
}

func (n *notificationService) MarkRead(id uint) error {
	notification, err := n.db.Notifications.Get(id)
	if err != nil {
		return err
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

func (n *notificationService) MarkReadMany(ids []uint) error {
	notifications, err := n.db.Notifications.GetMany(ids)
	if err != nil {
		return err
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

func (n *notificationService) MarkUnRead(id uint) error {
	return n.db.Notifications.MarkUnread(id)
}
