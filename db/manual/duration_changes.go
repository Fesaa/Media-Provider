package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

func SubscriptionDurationChanges(db *gorm.DB) error {
	if getCurrentVersion(db) != "" {
		return nil
	}

	var subscriptions []models.Subscription
	res := db.Find(&subscriptions)
	if res.Error != nil {
		return res.Error
	}

	for _, sub := range subscriptions {
		if sub.RefreshFrequency < 2 {
			sub.RefreshFrequency = 2
			if err := db.Save(&sub).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
