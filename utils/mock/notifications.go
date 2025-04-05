package mock

import "github.com/Fesaa/Media-Provider/db/models"

type Notifications struct {
	NotifyFunc          func(models.Notification)
	NotifyHelperFunc    func(string, string, string, models.NotificationColour, models.NotificationGroup)
	NotifyContentFunc   func(string, string, string, ...models.NotificationColour)
	NotifyContentQFunc  func(string, string, ...models.NotificationColour)
	NotifySecurityFunc  func(string, string, string, ...models.NotificationColour)
	NotifySecurityQFunc func(string, string, ...models.NotificationColour)
	NotifyGeneralFunc   func(string, string, string, ...models.NotificationColour)
	NotifyGeneralQFunc  func(string, string, ...models.NotificationColour)
	MarkReadFunc        func(uint) error
	MarkReadManyFunc    func([]uint) error
	MarkUnReadFunc      func(uint) error
}

func (m Notifications) Notify(notification models.Notification) {
	if m.NotifyFunc != nil {
		m.NotifyFunc(notification)
	}
}

func (m Notifications) NotifyHelper(title, summary, body string, colour models.NotificationColour, group models.NotificationGroup) {
	if m.NotifyHelperFunc != nil {
		m.NotifyHelperFunc(title, summary, body, colour, group)
	}
}

func (m Notifications) NotifyContent(title, summary, body string, colours ...models.NotificationColour) {
	if m.NotifyContentFunc != nil {
		m.NotifyContentFunc(title, summary, body, colours...)
	}
}

func (m Notifications) NotifyContentQ(title, body string, colours ...models.NotificationColour) {
	if m.NotifyContentQFunc != nil {
		m.NotifyContentQFunc(title, body, colours...)
	}
}

func (m Notifications) NotifySecurity(title, summary, body string, colours ...models.NotificationColour) {
	if m.NotifySecurityFunc != nil {
		m.NotifySecurityFunc(title, summary, body, colours...)
	}
}

func (m Notifications) NotifySecurityQ(title, body string, colours ...models.NotificationColour) {
	if m.NotifySecurityQFunc != nil {
		m.NotifySecurityQFunc(title, body, colours...)
	}
}

func (m Notifications) NotifyGeneral(title, summary, body string, colours ...models.NotificationColour) {
	if m.NotifyGeneralFunc != nil {
		m.NotifyGeneralFunc(title, summary, body, colours...)
	}
}

func (m Notifications) NotifyGeneralQ(title, body string, colours ...models.NotificationColour) {
	if m.NotifyGeneralQFunc != nil {
		m.NotifyGeneralQFunc(title, body, colours...)
	}
}

func (m Notifications) MarkRead(id uint) error {
	if m.MarkReadFunc != nil {
		return m.MarkReadFunc(id)
	}
	return nil
}

func (m Notifications) MarkReadMany(ids []uint) error {
	if m.MarkReadManyFunc != nil {
		return m.MarkReadManyFunc(ids)
	}
	return nil
}

func (m Notifications) MarkUnRead(id uint) error {
	if m.MarkUnReadFunc != nil {
		return m.MarkUnReadFunc(id)
	}
	return nil
}
