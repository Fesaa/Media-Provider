package mock

import "github.com/Fesaa/Media-Provider/db/models"

type Notifications struct {
}

func (m Notifications) Notify(notification models.Notification) {
}

func (m Notifications) NotifyHelper(title, summary, body string, colour models.NotificationColour, group models.NotificationGroup) {
}

func (m Notifications) NotifyContent(title, summary, body string, colours ...models.NotificationColour) {
}

func (m Notifications) NotifyContentQ(title, body string, colours ...models.NotificationColour) {
}

func (m Notifications) NotifySecurity(title, summary, body string, colours ...models.NotificationColour) {
}

func (m Notifications) NotifySecurityQ(title, body string, colours ...models.NotificationColour) {
}

func (m Notifications) NotifyGeneral(title, summary, body string, colours ...models.NotificationColour) {
}

func (m Notifications) NotifyGeneralQ(title, body string, colours ...models.NotificationColour) {
}

func (m Notifications) MarkRead(id uint) error {
	return nil
}

func (m Notifications) MarkUnRead(id uint) error {
	return nil
}
