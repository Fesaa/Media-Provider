package mock

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
)

type Notifications struct {
}

func (n Notifications) GetNotifications(ctx context.Context, user models.User, time time.Time) ([]models.Notification, error) {
	return []models.Notification{}, nil
}

func (n Notifications) Notify(ctx context.Context, notification models.Notification) {
}

func (n Notifications) MarkRead(ctx context.Context, user models.User, i int) error {
	return nil
}

func (n Notifications) MarkReadMany(ctx context.Context, user models.User, ints []int) error {
	return nil
}

func (n Notifications) MarkUnRead(ctx context.Context, user models.User, i int) error {
	return nil
}

func (n Notifications) Delete(ctx context.Context, user models.User, i int) error {
	return nil
}

func (n Notifications) DeleteMany(ctx context.Context, user models.User, ints []int) error {
	return nil
}
