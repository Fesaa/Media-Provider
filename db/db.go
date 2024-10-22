package db

import (
	"database/sql"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	_ "modernc.org/sqlite"
	"path"
)

var (
	DB *sql.DB
)

func Init() {
	var err error
	DB, err = sql.Open("sqlite", path.Join(config.Dir, "media-provider.db"))
	if err != nil {
		log.Fatal("failed to open a connection to the database", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("failed to ping the database", err)
	}

	if err = migrate(); err != nil {
		log.Fatal("failed to migrate", err)
	}

	log.Info("successfully connected to the database")
}

func Close() {
	if DB == nil {
		log.Warn("tried closing DB, while none was initialized")
		return
	}

	if err := DB.Close(); err != nil {
		log.Error("failed to close the DB", err)
	}
	DB = nil
}
