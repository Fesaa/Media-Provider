package config

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSecret(length int) (string, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(secret), nil
}

func (c *Config) RefreshApiKey(syncID int) error {
	panic("not implemented")
}

func (c *Config) Update(config Config, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	config.Version = c.Version
	config.Secret = c.Secret
	config.SyncId = syncID
	return Save(&config)
}
