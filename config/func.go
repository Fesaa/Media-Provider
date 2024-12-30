package config

func (current *Config) Update(config Config, syncID int) error {
	if current.SyncId != syncID {
		return InvalidSyncID
	}

	config.Version = current.Version
	config.Secret = current.Secret
	config.SyncId = syncID
	config.HasUpdatedDB = current.HasUpdatedDB
	return current.Save(&config, true)
}
