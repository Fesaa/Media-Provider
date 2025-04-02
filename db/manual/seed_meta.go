package manual

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
	"time"
)

func InitialMetadata(db *gorm.DB) error {
	var rows []models.MetadataRow
	if err := db.Find(&rows).Error; err != nil {
		return err
	}

	if len(rows) > 0 {
		return nil
	}

	rows = append(rows, models.MetadataRow{
		Key:   models.FirstInstalledVersion,
		Value: config.Version,
	})
	rows = append(rows, models.MetadataRow{
		Key:   models.InstallDate,
		Value: time.Now().Format(time.DateTime),
	})
	rows = append(rows, models.MetadataRow{
		Key:   models.InstalledVersion,
		Value: config.Version,
	})

	return db.Create(&rows).Error
}
