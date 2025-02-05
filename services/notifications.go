package services

import (
	"database/sql"
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

type NotificationService interface {
	NotifyStruct(models.Notification)
	// Notify constructs the struct, and calls NotifyStruct
	Notify(title, summary, body string, colour models.NotificationColour, group models.NotificationGroup)

	// NotifyContent calls Notify with GroupContent, default colour is blue(info)
	NotifyContent(title, summary, body string, colours ...models.NotificationColour)
	// NotifyContentQ calls NotifyContent with summary being the first 40 characters of body
	NotifyContentQ(title, body string, colours ...models.NotificationColour)

	// NotifySecurity calls Notify with GroupSecurity, default colour is orange(warn)
	NotifySecurity(title, summary, body string, colours ...models.NotificationColour)
	// NotifySecurityQ calls NotifySecurity with summary being the first 40 characters of body
	NotifySecurityQ(title, body string, colours ...models.NotificationColour)

	// NotifyGeneral calls Notify with GroupGeneral, default colour is white(secondary)
	NotifyGeneral(title, summary, body string, colours ...models.NotificationColour)
	// NotifyGeneralQ calls NotifyGeneral with summary being the first 40 characters of body
	NotifyGeneralQ(title, body string, colours ...models.NotificationColour)

	// MarkRead marks the notification with id as read, and sends the NotificationRead event through SignalR
	MarkRead(id uint) error
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

func (n *notificationService) NotifyStruct(notification models.Notification) {
	n.log.Debug().Any("notification", notification).Msg("adding notification")
	n.signalR.Notify(notification)
	if err := n.db.Notifications.New(notification); err != nil {
		n.log.Error().Err(err).Msg("unable to add notification")
	}
}

func (n *notificationService) Notify(title, summary, body string, colour models.NotificationColour, group models.NotificationGroup) {
	n.NotifyStruct(models.Notification{
		Title:   title,
		Summary: summary,
		Body:    body,
		Colour:  colour,
		Group:   group,
		Read:    false,
		ReadAt:  sql.NullTime{},
	})
}

func (n *notificationService) NotifyContent(title, summary, body string, colours ...models.NotificationColour) {
	colour := utils.OrDefault(colours, models.Blue)
	n.Notify(title, summary, body, colour, models.GroupContent)
}

func (n *notificationService) NotifyContentQ(title, body string, colours ...models.NotificationColour) {
	n.NotifyContent(title, utils.Shorten(body, SummarySize), body, colours...)
}

func (n *notificationService) NotifySecurity(title, summary, body string, colours ...models.NotificationColour) {
	colour := utils.OrDefault(colours, models.Orange)
	n.Notify(title, summary, body, colour, models.GroupSecurity)
}

func (n *notificationService) NotifySecurityQ(title, body string, colours ...models.NotificationColour) {
	n.NotifySecurity(title, utils.Shorten(body, SummarySize), body, colours...)
}

func (n *notificationService) NotifyGeneral(title, summary, body string, colours ...models.NotificationColour) {
	colour := utils.OrDefault(colours, models.White)
	n.Notify(title, summary, body, colour, models.GroupGeneral)
}

func (n *notificationService) NotifyGeneralQ(title, body string, colours ...models.NotificationColour) {
	n.NotifyGeneral(title, utils.Shorten(body, SummarySize), body, colours...)
}

func (n *notificationService) MarkRead(id uint) error {
	if err := n.db.Notifications.MarkRead(id); err != nil {
		return err
	}

	n.signalR.Broadcast(payload.EvenTypeNotificationRead, fiber.Map{})
	return nil
}

func (n *notificationService) MarkUnRead(id uint) error {
	notification, err := n.db.Notifications.Get(id)
	if err != nil {
		return err
	}
	if err = n.db.Notifications.MarkUnread(id); err != nil {
		return err
	}

	n.signalR.Notify(notification)
	return nil
}
