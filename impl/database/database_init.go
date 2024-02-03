package database

import "errors"

func (d *DatabaseImpl) loadTables() error {
	_, e := d.db.Exec(`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL
		);`)
	if e != nil {
		return errors.New("failed to create users table: " + e.Error())
	}

	_, e = d.db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires TIMESTAMP NOT NULL
		);`)
	if e != nil {
		return errors.New("failed to create tokens table: " + e.Error())
	}

	_, e = d.db.Exec(`CREATE TABLE IF NOT EXISTS roles (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			description VARCHAR(255) NOT NULL,
			value INTEGER NOT NULL
		);`)
	if e != nil {
		return errors.New("failed to create roles table: " + e.Error())
	}

	_, e = d.db.Exec(`CREATE TABLE IF NOT EXISTS user_roles (
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE
			);`)
	if e != nil {
		return errors.New("failed to create user_roles table: " + e.Error())
	}

	_, e = d.db.Exec(`CREATE TABLE IF NOT EXISTS permissions (
			key VARCHAR(255) UNIQUE NOT NULL,
			description VARCHAR(255) NOT NULL,
			value INTEGER NOT NULL
			);`)
	if e != nil {
		return errors.New("failed to create permissions table: " + e.Error())
	}

	return nil
}
