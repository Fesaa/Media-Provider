package db

import (
	"database/sql"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/api"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/log"
	_ "modernc.org/sqlite"
	"path"
)

type Database struct {
	db            *sql.DB
	Users         api.Users
	Pages         api.Pages
	Subscriptions api.Subscriptions
}

func (db *Database) DB() *sql.DB {
	return db.db
}

var (
	theDb *sql.DB
)

func Connect() (*Database, error) {
	var err error
	theDb, err = sql.Open("sqlite", path.Join(config.Dir, "media-provider.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = theDb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping a database connection: %w", err)
	}

	if err = migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Info("successfully connected to the database")
	return &Database{
		db:            theDb,
		Users:         models.NewUsers(theDb),
		Pages:         models.NewPages(theDb),
		Subscriptions: models.NewSubscriptions(theDb),
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
