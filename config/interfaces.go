package config

import (
	"os"
	"path"
)

func (c *Config) GetRootDir() string {
	return OrDefault(c.RootDir, path.Join(OrDefault(os.Getenv("CONFIG_DIR"), "."), "temp"))
}

func (c *Config) GetMaxConcurrentTorrents() int {
	return c.Downloader.MaxConcurrentTorrents
}

func (c *Config) GetMaxConcurrentMangadexImages() int {
	return c.Downloader.MaxConcurrentMangadexImages
}
