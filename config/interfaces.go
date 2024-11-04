package config

import (
	"math"
	"os"
	"path"
)

func (c *Config) GetRootDir() string {
	return OrDefault(c.RootDir, path.Join(OrDefault(os.Getenv("CONFIG_DIR"), "."), "temp"))
}

func (c *Config) GetMaxConcurrentTorrents() int {
	return int(math.Max(1, math.Min(10, float64(c.Downloader.MaxConcurrentTorrents))))
}

func (c *Config) GetMaxConcurrentMangadexImages() int {
	return c.GetMaxConcurrentImages()
}

func (c *Config) GetMaxConcurrentImages() int {
	return int(math.Max(1, math.Min(5, float64(c.Downloader.MaxConcurrentMangadexImages))))
}
