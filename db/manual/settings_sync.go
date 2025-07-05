package manual

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var defaults = []models.ServerSetting{
	{
		Key:   models.CacheType,
		Value: string(config.MEMORY),
	},
	{
		Key:   models.RedisAddr,
		Value: "",
	},
	{
		Key:   models.MaxConcurrentTorrents,
		Value: "5",
	},
	{
		Key:   models.MaxConcurrentImages,
		Value: "5",
	},
	{
		Key:   models.DisableIpv6,
		Value: "false",
	},
	{
		Key:   models.RootDir,
		Value: "",
	},
	{
		Key:   models.OidcAuthority,
		Value: "",
	},
	{
		Key:   models.OidcClientID,
		Value: "",
	},
	{
		Key:   models.OidcDisablePasswordLogin,
		Value: "false",
	},
	{
		Key:   models.OidcAutoLogin,
		Value: "false",
	},
}

func SyncSettings(db *gorm.DB, log zerolog.Logger) error {
	var ext []models.ServerSetting
	if err := db.Find(&ext).Error; err != nil {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, def := range defaults {
			ok := utils.Any(ext, func(setting models.ServerSetting) bool {
				return setting.Key == def.Key
			})
			if ok {
				continue
			}

			log.Debug().Any("key", def.Key).Str("value", def.Value).Msg("inserting default setting")
			if err := tx.Create(&def).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
