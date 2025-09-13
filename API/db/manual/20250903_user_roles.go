package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func UpdateUserRoles(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	defaultUser, err := getDefaultUser(ctx, db)
	if err != nil {
		return err
	}

	log.Debug().Str("user", defaultUser.Name).Msg("Adding all roles to original user")

	defaultUser.Roles = models.AllRoles

	return db.WithContext(ctx).Save(&defaultUser).Error
}
