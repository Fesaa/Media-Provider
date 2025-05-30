package manual

import (
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"strings"
)

func RemoveAllDeleted(db *gorm.DB, log zerolog.Logger) error {
	if getCurrentVersion(db) != "" {
		log.Trace().Msg("Skipping changes, Media-Provider installed after changes are needed")
		return nil
	}
	for _, table := range []string{"modifier_values", "modifiers", "notifications", "pages", "password_resets", "preferences", "subscription_infos", "subscriptions", "tags", "users"} {
		res := db.Exec("DELETE FROM `" + table + "` WHERE deleted_at IS NOT NULL")
		if res.Error != nil {
			// Migration is running after these columns were removed
			if strings.Contains(res.Error.Error(), "no such column") {
				return nil
			}
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
