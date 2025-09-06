package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func SubscriptionDurationChanges(db *gorm.DB, log zerolog.Logger) error {
	if getCurrentVersion(db) != "" {
		log.Trace().Msg("Skipping changes, Media-Provider installed after changes are needed")
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
