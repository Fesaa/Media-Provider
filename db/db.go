package db

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/impl"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"path"
)

type Database struct {
	db            *gorm.DB
	Users         models.Users
	Pages         models.Pages
	Subscriptions models.Subscriptions
	Preferences   models.Preferences
}

func (db *Database) DB() *gorm.DB {
	return db.db
}

func DatabaseProvider(log zerolog.Logger) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(path.Join(config.Dir, "media-provider.db")), &gorm.Config{
		Logger:               gormLogger(log),
		FullSaveAssociations: true,
	})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(models.MODELS...); err != nil {
		return nil, err
	}

	if err = manualMigration(db, log); err != nil {
		return nil, err
	}

	return &Database{
		db:            db,
		Users:         impl.Users(db),
		Pages:         impl.Pages(db),
		Subscriptions: impl.Subscriptions(db),
		Preferences:   impl.Preferences(db),
	}, nil
}
