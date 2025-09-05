package db

import (
	"testing"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func databaseHelper(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"))
	if err != nil {
		t.Fatal(err)
	}

	if err = db.AutoMigrate(models.MODELS...); err != nil {
		t.Fatal(err)
	}

	return db
}
