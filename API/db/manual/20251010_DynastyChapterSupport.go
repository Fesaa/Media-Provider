package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func MigrateDynastySubscriptionIds(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	var subs []models.Subscription
	if err := db.WithContext(ctx).Find(&subs).Error; err != nil {
		return err
	}

	var toUpdate []models.Subscription

	for _, sub := range subs {
		if sub.Provider != models.DYNASTY {
			continue
		}

		toUpdate = append(toUpdate, sub)
	}

	if len(toUpdate) == 0 {
		return nil
	}

	log.Info().Int("count", len(toUpdate)).Msg("Updating dynasty subscription ids")

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sub := range toUpdate {
			sub.ContentId = "/series/" + sub.ContentId

			if err := tx.Save(&sub).Error; err != nil {
				return err
			}
		}

		return nil
	})

}
