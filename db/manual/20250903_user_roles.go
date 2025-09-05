package manual

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func UpdateUserRoles(db *gorm.DB, log zerolog.Logger) error {
	var defaultUser models.User
	if err := db.Find(&defaultUser, "original = ?", true).Error; err != nil {
		return err
	}

	log.Debug().Str("user", defaultUser.Name).Msg("Adding all roles to original user")

	defaultUser.Roles = models.AllRoles

	return db.Save(&defaultUser).Error
}
