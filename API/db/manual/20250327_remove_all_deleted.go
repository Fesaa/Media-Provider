package manual

import (
	"context"
	"strings"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func RemoveAllDeleted(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	if getCurrentVersion(ctx, db) != "" {
		log.Trace().Msg("Skipping changes, Media-Provider installed after changes are needed")
		return nil
	}
	for _, table := range []string{"modifier_values", "modifiers", "notifications", "pages", "password_resets", "preferences", "subscription_infos", "subscriptions", "tags", "users"} {
		res := db.WithContext(ctx).Exec("DELETE FROM `" + table + "` WHERE deleted_at IS NOT NULL")
		if res.Error != nil {
			// Migration is running after these columns were removed
			if strings.Contains(res.Error.Error(), "no such column") {
				return nil
			}
			return res.Error
		}

		res = db.WithContext(ctx).Exec("DROP INDEX `idx_" + table + "_deleted_at`")
		if res.Error != nil {
			return res.Error
		}

		res = db.WithContext(ctx).Exec("ALTER TABLE `" + table + "`DROP COLUMN deleted_at;")
		if res.Error != nil {
			return res.Error
		}
	}
	return nil
}
