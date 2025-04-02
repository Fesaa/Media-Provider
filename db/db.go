package db

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/impl"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"gorm.io/gorm"
	"os"
	"path"
)

type Database struct {
	db            *gorm.DB
	Users         models.Users
	Pages         models.Pages
	Subscriptions models.Subscriptions
	Preferences   models.Preferences
	Notifications models.Notifications
	Metadata      models.Metadata
}

func (db *Database) DB() *gorm.DB {
	return db.db
}

func DatabaseProvider(log zerolog.Logger) (*Database, error) {
	dsn := config.OrDefault(os.Getenv("DATABASE_DSN"), path.Join(config.Dir, "media-provider.db"))
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

	return &Database{
		db:            db,
		Users:         impl.Users(db),
		Pages:         impl.Pages(db),
		Subscriptions: impl.Subscriptions(db),
		Preferences:   impl.Preferences(db),
		Notifications: impl.Notifications(db),
		Metadata:      impl.Metadata(db),
	}, nil
}

func ModelsProvider(db *Database, c *dig.Container) {
	utils.Must(c.Provide(utils.Identity(db.Users)))
	utils.Must(c.Provide(utils.Identity(db.Pages)))
	utils.Must(c.Provide(utils.Identity(db.Subscriptions)))
	utils.Must(c.Provide(utils.Identity(db.Preferences)))
	utils.Must(c.Provide(utils.Identity(db.Notifications)))
	utils.Must(c.Provide(utils.Identity(db.Metadata)))
}
