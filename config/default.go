package config

import (
	"log/slog"
	"os"
	"path"
)

func defaultConfig() *Config {
	secret, err := GenerateSecret(64)
	if err != nil {
		panic(err)
	}

	return &Config{
		SyncId:  0,
		RootDir: path.Join(OrDefault(os.Getenv("CONFIG_DIR"), "."), "temp"),
		BaseUrl: "",
		Secret:  secret,
		Cache: CacheConfig{
			Type: MEMORY,
		},
		Logging: Logging{
			Level:   slog.LevelInfo,
			Source:  true,
			Handler: LogHandlerText,
		},
		Downloader: Downloader{
			MaxConcurrentTorrents:       5,
			MaxConcurrentMangadexImages: 4,
		},
	}
}
