package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

func getCurrentVersion(ctx context.Context, db *gorm.DB) string {
	var row models.ServerSetting
	err := db.WithContext(ctx).First(&row, "key = ?", models.InstalledVersion).Error
	if err != nil {
		return ""
	}
	return row.Value
}

func getDefaultUser(ctx context.Context, db *gorm.DB) (models.User, error) {
	var defaultUser models.User
	err := db.WithContext(ctx).Find(&defaultUser, "original = ?", true).Error
	return defaultUser, err
}
