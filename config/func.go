package config

func (c *Config) Update(config Config, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	config.Version = c.Version
	config.Secret = c.Secret
	config.SyncId = syncID
	return Save(&config)
}
