package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func RemoveRedirectUrl(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	return db.WithContext(ctx).Delete(&models.ServerSetting{}, "key = ?", 12).Error
}
