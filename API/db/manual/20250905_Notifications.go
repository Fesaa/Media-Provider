package manual

import (
	"database/sql"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func AssignNotificationsToDefaultUser(db *gorm.DB, log zerolog.Logger) error {
	defaultUser, err := getDefaultUser(db)
	if err != nil {
		return err
	}

	var notifications []models.Notification
	if err = db.Find(&notifications).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, notification := range notifications {
			notification.Owner = sql.NullInt32{Int32: int32(defaultUser.ID), Valid: true}
			if err = tx.Save(&notification).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
