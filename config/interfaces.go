package config

func (c *Config) GetRootDir() string {
	return c.RootDir
}

func (c *Config) GetMaxConcurrentTorrents() int {
	return c.Downloader.MaxConcurrentTorrents
}

func (c *Config) GetMaxConcurrentMangadexImages() int {
	return c.Downloader.MaxConcurrentMangadexImages
}
