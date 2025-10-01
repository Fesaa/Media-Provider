package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/db/repository"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func MigrateMetadataToSettings(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	if !ensureTablesExist(ctx, db, "metadata_rows") {
		return nil
	}

	type row struct {
		Key   int
		Value string
	}

	var rows []row
	err := db.WithContext(ctx).Table("metadata_rows").Find(&rows).Error
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch metadata rows")
		return err
	}

	settingRows := make([]models.ServerSetting, 0)

	for _, metadataRow := range rows {
		switch metadataRow.Key {
		case 0:
			settingRows = append(settingRows, models.ServerSetting{
				Key:   models.InstalledVersion,
				Value: metadataRow.Value,
			})
		case 1:
			settingRows = append(settingRows, models.ServerSetting{
				Key:   models.FirstInstalledVersion,
				Value: metadataRow.Value,
			})
		case 2:
			settingRows = append(settingRows, models.ServerSetting{
				Key:   models.InstallDate,
				Value: metadataRow.Value,
			})
		}
	}

	repo := repository.NewSettingsRepository(db)

	return repo.Update(ctx, settingRows)
}
