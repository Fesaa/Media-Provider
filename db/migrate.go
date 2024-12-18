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
		`ALTER TABLE pages ADD COLUMN sortValue INTEGER NOT NULL DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN original BIT NOT NULL DEFAULT 0`,
		`CREATE TABLE password_reset (
	user_id INTEGER REFERENCES users(id),
	key TEXT NOT NULL,
	expiry BIGINT NOT NULL
);`,
		`
CREATE TABLE subscriptions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	provider INTEGER NOT NULL,
	contentId INTEGER NOT NULL,
	refreshFrequency INTEGER NOT NULL
);

CREATE TABLE subscription_info (
	subscription_id INTEGER REFERENCES subscriptions(id),
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	lastCheck TIMESTAMP NOT NULL,
	lastCheckSuccess BOOLEAN NOT NULL
);
`,
		`ALTER TABLE subscription_info ADD COLUMN baseDir TEXT NOT NULL DEFAULT '';`,
	}
)

func migrate() error {
	if _, err := theDb.Exec(migrationTable); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	for index, m := range migrations {
		log.Trace("checking migration", slog.Int("idx", index))

		row := theDb.QueryRow("SELECT executed FROM migrations WHERE idx = ?;", index)
		var executed bool
		if err := row.Scan(&executed); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("could not check for migration %d: %w", index, err)
			}
		}
		if executed {
			log.Trace("skipping migration, already executed", slog.Int("idx", index))
			continue
		}

		if _, err := theDb.Exec(m); err != nil {
			return fmt.Errorf("could not execute migration %d: %w\n%s", index, err, m)
		}
		if _, err := theDb.Exec("INSERT INTO migrations (idx, executed) VALUES (?, true);", index); err != nil {
			return fmt.Errorf("could not save migration %d: %w", index, err)
		}

		log.Info("successfully executed migration", slog.Int("idx", index))
		if log.IsTraceEnabled() {
			log.Trace("executed migrations", slog.Int("idx", index), slog.String("query", m))
		}
	}

	return nil
}
