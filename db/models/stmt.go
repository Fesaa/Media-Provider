package models

import "database/sql"

func Init(db *sql.DB) error {
	var err error

	if err = initUser(db); err != nil {
		return err
	}

	return nil
}
