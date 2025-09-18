package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func AssignSubscriptionsToUser(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	defaultUser, err := getDefaultUser(ctx, db)
	if err != nil {
		return err
	}

	var subs []models.Subscription
	if err = db.WithContext(ctx).Find(&subs).Error; err != nil {
		return err
	}

	if len(subs) == 0 {
		return nil
	}

	log.Debug().Str("user", defaultUser.Name).Msg("Assigning subscriptions to user")

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sub := range subs {
			sub.Owner = defaultUser.ID
			if err = tx.Save(&sub).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
