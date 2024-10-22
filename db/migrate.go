package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/log"
	"log/slog"
)

var (
	migrationTable = `CREATE TABLE IF NOT EXISTS migrations (
    idx INT NOT NULL,
    executed BOOL NOT NULL
)`
	migrations = []string{
		`CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    apiKey TEXT NOT NULL
);`,
		`ALTER TABLE users ADD COLUMN permission INTEGER`,
		`CREATE TABLE pages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	customRootDir TEXT NOT NULL
);`,
		`CREATE TABLE providers (
	page_id INTEGER REFERENCES pages(id),
	provider INTEGER NOT NULL
);`,
		`
CREATE TABLE dirs (
	page_id INTEGER REFERENCES pages(id),
	dir TEXT NOT NULL
);`,
		`CREATE TABLE modifiers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	page_id INTEGER REFERENCES pages(id),
	title TEXT NOT NULL,
	type INTEGER NOT NULL,
	key TEXT NOT NULL
);`,
		`CREATE TABLE modifier_values (
	modifier_id INTEGER REFERENCES modifiers(id),
	key TEXT NOT NULL,
	value TEXT NOT NULL
);`,
	}
)

func migrate() error {
	if _, err := DB.Exec(migrationTable); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	for index, m := range migrations {
		log.Trace("checking migration", slog.Int("idx", index))

		row := DB.QueryRow("SELECT executed FROM migrations WHERE idx = ?;", index)
		var executed bool
		if err := row.Scan(&executed); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("could not check for migration %d: %w", index, err)
			}
		}
		if executed {
			log.Debug("skipping migration, already executed", slog.Int("idx", index))
			continue
		}

		if _, err := DB.Exec(m); err != nil {
			return fmt.Errorf("could not execute migration %d: %w\n%s", index, err, m)
		}
		if _, err := DB.Exec("INSERT INTO migrations (idx, executed) VALUES (?, true);", index); err != nil {
			return fmt.Errorf("could not save migration %d: %w", index, err)
		}

		log.Info("successfully executed migration", slog.Int("idx", index))
		if log.IsTraceEnabled() {
			log.Trace("executed migrations", slog.Int("idx", index), slog.String("query", m))
		}
	}

	return nil
}
