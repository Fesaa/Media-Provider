package db

import (
	"database/sql"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/impl"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/log"
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

var (
	theDb *sql.DB
)

func Connect() (*Database, error) {
	db, err := gorm.Open(sqlite.Open(path.Join(config.Dir, "media-provider.db")), &gorm.Config{
		Logger:               logger.Default.LogMode(logger.Warn),
		FullSaveAssociations: true,
	})
	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}

	if err = db.AutoMigrate(models.MODELS...); err != nil {
		log.Error("failed to auto migrate", "err", err)
		panic(err)
	}

	return &Database{
		db:            db,
		Users:         impl.Users(db),
		Pages:         impl.Pages(db),
		Subscriptions: impl.Subscriptions(db),
	}, nil
}

func Close() {
	if theDb == nil {
		log.Warn("tried closing theDb, while none was initialized")
		return
	}

	if err := theDb.Close(); err != nil {
		log.Error("failed to close the theDb", err)
	}
	theDb = nil
}
