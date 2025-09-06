package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func RemoveRedirectUrl(db *gorm.DB, log zerolog.Logger) error {
	return db.Delete(&models.ServerSetting{}, "key = ?", 12).Error
}
