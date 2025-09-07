package manual

import (
	"context"
	"database/sql"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func AssignNotificationsToDefaultUser(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	defaultUser, err := getDefaultUser(ctx, db)
	if err != nil {
		return err
	}

	var notifications []models.Notification
	if err = db.WithContext(ctx).Find(&notifications).Error; err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, notification := range notifications {
			notification.Owner = sql.NullInt32{Int32: int32(defaultUser.ID), Valid: true}
			if err = tx.Save(&notification).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
