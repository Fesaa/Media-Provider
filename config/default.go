package config

import (
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"os"
)

func DefaultConfig() *Config {
	secret, err := utils.GenerateSecret(64)
	if err != nil {
		panic(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &Config{
		RootDir:      orDefault(os.Getenv("CONFIG_DIR"), pwd),
		Secret:       secret,
		HasUpdatedDB: false,
		Logging: Logging{
			Level:   zerolog.InfoLevel,
			Source:  true,
			Handler: LogHandlerText,
		},
		Downloader: Downloader{
			MaxConcurrentTorrents:       5,
			MaxConcurrentMangadexImages: 5,
			DisableIpv6:                 false,
		},
		Cache: CacheConfig{
			Type: MEMORY,
		},
	}
}
