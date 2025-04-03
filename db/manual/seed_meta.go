package manual

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"time"
)

func InitialMetadata(db *gorm.DB, log zerolog.Logger) error {
	var rows []models.MetadataRow
	if err := db.Find(&rows).Error; err != nil {
		return err
	}

	if len(rows) > 0 {
		log.Trace().Msg("Metadata rows found, no need to set initial values. How did this happen?")
		return nil
	}

	rows = append(rows, models.MetadataRow{
		Key:   models.FirstInstalledVersion,
		Value: config.Version.String(),
	})
	rows = append(rows, models.MetadataRow{
		Key:   models.InstallDate,
		Value: time.Now().Format(time.DateTime),
	})
	rows = append(rows, models.MetadataRow{
		Key:   models.InstalledVersion,
		Value: config.Version.String(),
	})

	return db.Create(&rows).Error
}
