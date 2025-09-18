package manual

import (
	"context"
	"strings"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func RemoveRedirectUrl(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	var setting models.ServerSetting
	if err := db.WithContext(ctx).Where("key = ?", 12).First(&setting).Error; err != nil {
		return allowNoRecord(err)
	}

	if !strings.HasPrefix(setting.Value, "http") {
		return nil
	}

	return db.WithContext(ctx).Delete(&models.ServerSetting{}, "key = ?", 12).Error
}
