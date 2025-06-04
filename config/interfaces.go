package config

import (
	"math"
	"os"
	"path"
)

func (current *Config) GetRootDir() string {
	return orDefault(current.RootDir, path.Join(orDefault(os.Getenv("CONFIG_DIR"), "."), "temp"))
}

func (current *Config) GetMaxConcurrentTorrents() int {
	return int(math.Max(1, math.Min(10, float64(current.Downloader.MaxConcurrentTorrents))))
}

func (current *Config) GetMaxConcurrentMangadexImages() int {
	return current.GetMaxConcurrentImages()
}

func (current *Config) GetMaxConcurrentImages() int {
	return int(math.Max(1, math.Min(5, float64(current.Downloader.MaxConcurrentMangadexImages))))
}
