package manual

import (
	"gorm.io/gorm"
)

func RemoveAllDeleted(db *gorm.DB) error {
	for _, table := range []string{"modifier_values", "modifiers", "notifications", "pages", "password_resets", "preferences", "subscription_infos", "subscriptions", "tags", "users"} {
		res := db.Exec("DELETE FROM `" + table + "` WHERE deleted_at IS NOT NULL")
		if res.Error != nil {
			return res.Error
		}

		res = db.Exec("DROP INDEX `idx_" + table + "_deleted_at`")
		if res.Error != nil {
			return res.Error
		}

		res = db.Exec("ALTER TABLE `" + table + "`DROP COLUMN deleted_at;")
		if res.Error != nil {
			return res.Error
		}
	}
	return nil
}
