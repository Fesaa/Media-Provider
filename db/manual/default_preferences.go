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

func InsertEmptyArray(db *gorm.DB) error {
	var cur models.Preference
	res := db.First(&cur)
	if res.Error != nil {
		return res.Error
	}

	cur.DynastyGenreTags = []string{}
	return db.Save(&cur).Error
}
