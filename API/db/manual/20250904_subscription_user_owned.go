package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func AssignSubscriptionsToUser(db *gorm.DB, log zerolog.Logger) error {
	defaultUser, err := getDefaultUser(db)
	if err != nil {
		return err
	}

	log.Debug().Str("user", defaultUser.Name).Msg("Assigning subscriptions to user")

	var subs []models.Subscription
	if err := db.Find(&subs).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, sub := range subs {
			sub.Owner = defaultUser.ID
			if err := tx.Save(&sub).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
