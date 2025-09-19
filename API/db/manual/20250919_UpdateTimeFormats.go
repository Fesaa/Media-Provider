package manual

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func UpdateTimeFormats(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	timeKeys := []models.SettingKey{models.InstallDate, models.LastUpdateDate}

	for _, key := range timeKeys {
		if err := updateTimeFormatForKey(ctx, db, key); err != nil {
			return fmt.Errorf("UpdateTimeFormats: failed to update %v: %w", key, err)
		}
	}

	return nil
}

func updateTimeFormatForKey(ctx context.Context, db *gorm.DB, key models.SettingKey) error {
	var setting models.ServerSetting
	err := db.WithContext(ctx).Where(&models.ServerSetting{Key: key}).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	t, err := time.Parse(time.DateTime, setting.Value)
	if err != nil {
		return nil
	}

	setting.Value = t.Format(time.RFC3339)

	return db.WithContext(ctx).Where(&models.ServerSetting{Key: key}).Save(&setting).Error
}
