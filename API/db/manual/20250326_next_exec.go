package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func SubscriptionNextExec(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	if getCurrentVersion(ctx, db) != "" {
		log.Trace().Msg("Skipping changes, Media-Provider installed after changes are needed")
		return nil
	}
	var subscriptions []models.Subscription
	res := db.WithContext(ctx).Preload("Info").Find(&subscriptions)
	if res.Error != nil {
		return res.Error
	}

	var pref models.Preference
	res = db.WithContext(ctx).Find(&pref)
	if res.Error != nil {
		return res.Error
	}

	for _, sub := range subscriptions {
		sub.Info.NextExecution = sub.NextExecution(pref.SubscriptionRefreshHour)
		if err := db.WithContext(ctx).Save(&sub).Error; err != nil {
			return err
		}
	}

	return nil
}
