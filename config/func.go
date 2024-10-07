package config

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

func GenerateSecret(length int) (string, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(secret), nil
}

func ApiKey() (string, error) {
	bytes := make([]byte, 16)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	apiKey := hex.EncodeToString(bytes)
	return apiKey, nil
}

func (c *Config) RefreshApiKey(syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	apiKey, err := ApiKey()
	if err != nil {
		slog.Error("could not refresh api key", "err", err)
		return fiber.ErrInternalServerError
	}

	c.ApiKey = apiKey
	return c.Save()
}

func (c *Config) Update(config Config, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	config.Version = c.Version
	config.Secret = c.Secret
	config.ApiKey = c.ApiKey
	config.SyncId = syncID
	config.Pages = c.Pages
	return Save(&config)
}

func (c *Config) RemovePage(index, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	if index < 0 || index >= len(c.Pages) {
		slog.Debug("Invalid index", "index", index)
		return errors.New("invalid index")
	}

	c.Pages = append(c.Pages[:index], c.Pages[index+1:]...)
	return c.Save(true)
}

func (c *Config) AddPage(page Page, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}
	c.Pages = append(c.Pages, page)
	return c.Save()
}

func (c *Config) UpdatePage(page Page, index, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	c.Pages[index] = page
	return c.Save(true)
}

func (c *Config) MovePage(oldIndex, newIndex, syncID int) error {
	if c.SyncId != syncID {
		return InvalidSyncID
	}

	if oldIndex < 0 || oldIndex >= len(c.Pages) || newIndex < 0 || newIndex >= len(c.Pages) {
		slog.Debug("Invalid index", "old", oldIndex, "new", newIndex)
		return errors.New("invalid index")
	}

	page := c.Pages[oldIndex]
	c.Pages = append(c.Pages[:oldIndex], c.Pages[oldIndex+1:]...)
	c.Pages = append(c.Pages[:newIndex], append([]Page{page}, c.Pages[newIndex:]...)...)
	return c.Save()
}
