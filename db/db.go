package db

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/impl"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path"
)

type Database struct {
	db            *gorm.DB
	Users         models.Users
	Pages         models.Pages
	Subscriptions models.Subscriptions
}

func Connect() (*Database, error) {
	db, err := gorm.Open(sqlite.Open(path.Join(config.Dir, "media-provider.db")), &gorm.Config{
		Logger:               logger.Default.LogMode(logger.Warn),
		FullSaveAssociations: true,
	})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(models.MODELS...); err != nil {
		return nil, err
	}

	return &Database{
		db:            db,
		Users:         impl.Users(db),
		Pages:         impl.Pages(db),
		Subscriptions: impl.Subscriptions(db),
	}, nil
}
