package api

import "github.com/Fesaa/Media-Provider/db/models"

type Subscriptions interface {
	All() ([]models.Subscription, error)
	Get(int64) (*models.Subscription, error)

	New(models.Subscription) (*models.Subscription, error)
	Update(models.Subscription) error
	Delete(int64) error
}
