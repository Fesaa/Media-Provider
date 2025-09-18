package manual

import (
	"context"
	"errors"

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

// ensureTablesExist returns true if all tables exist
func ensureTablesExist(ctx context.Context, db *gorm.DB, tables ...string) bool {
	for _, table := range tables {
		if !db.WithContext(ctx).Migrator().HasTable(table) {
			return false
		}
	}

	return true
}

func allowNoRecord(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
