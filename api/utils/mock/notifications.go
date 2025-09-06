package mock

import (
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
)

type Notifications struct {
}

func (n Notifications) GetNotifications(user models.User, time time.Time) ([]models.Notification, error) {
	return []models.Notification{}, nil
}

func (n Notifications) Notify(notification models.Notification) {
}

func (n Notifications) MarkRead(user models.User, u int) error {
	return nil
}

func (n Notifications) MarkReadMany(user models.User, uints []int) error {
	return nil
}

func (n Notifications) MarkUnRead(user models.User, u int) error {
	return nil
}

func (n Notifications) Delete(user models.User, u int) error {
	return nil
}

func (n Notifications) DeleteMany(user models.User, uints []int) error {
	return nil
}
