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
		SyncId:  0,
		RootDir: orDefault(os.Getenv("CONFIG_DIR"), pwd),
		BaseUrl: "",
		Secret:  secret,
		Cache: CacheConfig{
			Type: MEMORY,
		},
		Logging: Logging{
			Level:   zerolog.InfoLevel,
			Source:  true,
			Handler: LogHandlerText,
		},
		Downloader: Downloader{
			MaxConcurrentTorrents:       5,
			MaxConcurrentMangadexImages: 4,
		},
	}
}
