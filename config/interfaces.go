package config

func (cfg *Config) GetRootDir() string {
	return cfg.RootDir
}

func (cfg *Config) GetMaxConcurrentTorrents() int {
	return cfg.Downloader.MaxConcurrentTorrents
}

func (cfg *Config) GetMaxConcurrentMangadexImages() int {
	return cfg.Downloader.MaxConcurrentMangadexImages
}
