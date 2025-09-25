package manual

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/metadata"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func InitialMetadata(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	var rows []models.ServerSetting
	if err := db.WithContext(ctx).Find(&rows).Error; err != nil {
		return err
	}

	if len(rows) > 0 {
		log.Trace().Msg("Metadata rows found, no need to set initial values. How did this happen?")
		return nil
	}

	rows = append(rows, models.ServerSetting{
		Key:   models.FirstInstalledVersion,
		Value: metadata.Version.String(),
	})
	rows = append(rows, models.ServerSetting{
		Key:   models.InstallDate,
		Value: time.Now().Format(time.RFC3339),
	})
	rows = append(rows, models.ServerSetting{
		Key:   models.InstalledVersion,
		Value: metadata.Version.String(),
	})

	return db.WithContext(ctx).Create(&rows).Error
}
