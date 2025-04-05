package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

func getCurrentVersion(db *gorm.DB) string {
	var row models.MetadataRow
	err := db.First(&row, "key = ?", models.InstalledVersion).Error
	if err != nil {
		return ""
	}
	return row.Value
}
