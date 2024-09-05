package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"
)

var (
	configDir     string
	configPath    string
	cfgLock       = sync.RWMutex{}
	InvalidSyncID = errors.New("invalid sync id")
)

func init() {
	file := OrDefault(os.Getenv("CONFIG_FILE"), "config.json")
	configDir = OrDefault(os.Getenv("CONFIG_DIR"), ".")
	configPath = path.Join(configDir, file)

	backupDir := path.Join(configDir, "backup")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err = os.Mkdir(backupDir, os.ModePerm); err != nil {
			slog.Warn("Failed to create missing backup directory... backups will fail. Please check permissions", "dir", backupDir)
		}
	}
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

func Save(cfg *Config, backUp ...bool) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	if current != nil {
		if current.SyncId != cfg.SyncId {
			return InvalidSyncID
		}
	}
	cfg.SyncId = cfg.SyncId + 1

	if len(backUp) > 0 && backUp[0] {
		backUpPath := path.Join(configDir, "backup", fmt.Sprintf("%d_config.json", time.Now().UTC().Unix()))
		slog.Info("Backing up config", "sync_id", cfg.SyncId, "to", backUpPath)
		if err := os.Rename(configPath, backUpPath); err != nil {
			slog.Error("Failed to backup config file", "path", backUpPath, "err", err)
		}
	}

	err := write(configPath, cfg)
	if err == nil {
		slog.SetLogLoggerLevel(cfg.Logging.Level)
	}
	return err
}

func (c *Config) Save(backUp ...bool) error {
	return Save(c, backUp...)
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
