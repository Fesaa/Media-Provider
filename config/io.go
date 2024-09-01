package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path"
	"sync"
)

var (
	configPath    string
	cfgLock       = sync.RWMutex{}
	InvalidSyncID = errors.New("invalid sync id")
)

func init() {
	file := OrDefault(os.Getenv("CONFIG_FILE"), "config.json")
	configPath = path.Join("", file)
}

func Load() (*Config, error) {
	cfgLock.RLock()
	defer cfgLock.RUnlock()

	cfg, err := read(configPath)

	if errors.Is(err, os.ErrNotExist) {
		slog.Warn("Config file not found, creating new one", "path", configPath)
		cfg = defaultConfig()
		err = write(configPath, cfg)
	}

	if err != nil {
		return nil, err
	}

	current = cfg
	return cfg, nil
}

func Save(cfg *Config) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	if current != nil {
		if current.SyncId != cfg.SyncId {
			return InvalidSyncID
		}
	}
	cfg.SyncId = cfg.SyncId + 1
	return write(configPath, cfg)
}

func (c *Config) Save() error {
	return Save(c)
}

func write(path string, cfg *Config) error {
	slog.Debug("Writing config", "path", path)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	current = cfg
	return nil
}

func read(path string) (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
