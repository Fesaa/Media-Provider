package manual

import (
	"context"
	"errors"
	"strings"

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

func allowNoTable(err error) error {
	if strings.Contains(err.Error(), "no such table:") {
		return nil
	}
	return err
}

func allowNoRecord(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
