package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

func InsertDefaultPreferences(db *gorm.DB) error {
	pref := models.Preference{
		SubscriptionRefreshHour: 0,
	}
	return db.Save(&pref).Error
}
