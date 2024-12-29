package models

import "gorm.io/gorm"

type ManualMigration struct {
	gorm.Model

	Success bool
	Name    string
}
