package db

import (
	"path"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/manual"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func DatabaseProvider(log zerolog.Logger) (*gorm.DB, error) {
	dsn := utils.OrElse(config.DatabaseDsn, path.Join(config.Dir, "media-provider.db"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:               gormLogger(log),
		FullSaveAssociations: true,
	})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(models.MODELS...); err != nil {
		return nil, err
	}

	if err = manualMigration(db, log.With().Str("handler", "migrations").Logger()); err != nil {
		return nil, err
	}

	if err = manual.SyncSettings(db, log.With().Str("handler", "settings").Logger()); err != nil {
		return nil, err
	}

	return db, nil
}
