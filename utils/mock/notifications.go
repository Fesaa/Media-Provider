package mock

import "github.com/Fesaa/Media-Provider/db/models"

type Notifications struct {
	NotifyFunc       func(models.Notification)
	MarkReadFunc     func(uint) error
	MarkReadManyFunc func([]uint) error
	MarkUnReadFunc   func(uint) error
}

func (m Notifications) Notify(notification models.Notification) {
	if m.NotifyFunc != nil {
		m.NotifyFunc(notification)
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
