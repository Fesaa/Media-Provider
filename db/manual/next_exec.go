package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

func SubscriptionNextExec(db *gorm.DB) error {
	var subscriptions []models.Subscription
	res := db.Preload("Info").Find(&subscriptions)
	if res.Error != nil {
		return res.Error
	}

	var pref models.Preference
	res = db.Find(&pref)
	if res.Error != nil {
		return res.Error
	}

	for _, sub := range subscriptions {
		sub.Info.NextExecution = sub.NextExecution(pref.SubscriptionRefreshHour)
		if err := db.Save(&sub).Error; err != nil {
			return err
		}
	}

	return nil
}
