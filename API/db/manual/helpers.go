package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

func getCurrentVersion(db *gorm.DB) string {
	var row models.ServerSetting
	err := db.First(&row, "key = ?", models.InstalledVersion).Error
	if err != nil {
		return ""
	}
	return row.Value
}

func getDefaultUser(db *gorm.DB) (models.User, error) {
	var defaultUser models.User
	err := db.Find(&defaultUser, "original = ?", true).Error
	return defaultUser, err
}
