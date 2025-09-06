package manual

import (
	"strconv"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func MigrateSettings(db *gorm.DB, log zerolog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	settings := []models.ServerSetting{
		{
			Key:   models.CacheType,
			Value: string(cfg.Cache.Type),
		},
		{
			Key:   models.RedisAddr,
			Value: cfg.Cache.RedisAddr,
		},
		{
			Key:   models.MaxConcurrentTorrents,
			Value: strconv.Itoa(cfg.Downloader.MaxConcurrentTorrents),
		},
		{
			Key:   models.MaxConcurrentImages,
			Value: strconv.Itoa(cfg.Downloader.MaxConcurrentMangadexImages),
		},
		{
			Key:   models.DisableIpv6,
			Value: strconv.FormatBool(cfg.Downloader.DisableIpv6),
		},
		{
			Key:   models.RootDir,
			Value: cfg.RootDir,
		},
	}

	for _, setting := range settings {
		if err = db.Create(&setting).Error; err != nil {
			return err
		}
	}

	return nil
}
