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
	Dir              string
	configPath       string
	cfgLock          = sync.RWMutex{}
	ErrInvalidSyncID = errors.New("invalid sync id")
)

func init() {
	file := OrDefault(os.Getenv("CONFIG_FILE"), "config.json")
	Dir = OrDefault(os.Getenv("CONFIG_DIR"), ".")
	configPath = path.Join(Dir, file)

	backupDir := path.Join(Dir, "backup")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err = os.Mkdir(backupDir, os.ModePerm); err != nil {
			slog.Warn("Failed to create missing backup directory... backups will fail. Please check permissions", "dir", backupDir)
		}
	}
}

func Load() (*Config, error) {

	cfg, err := read(configPath)

	if errors.Is(err, os.ErrNotExist) {
		slog.Warn("Config file not found, creating new one", "path", configPath)
		cfg = defaultConfig()
		err = write(configPath, cfg)
	}

	if err != nil {
		return nil, err
	}

	updatedConfig := update(*cfg)
	return &updatedConfig, nil
}

func (current *Config) Save(cfg *Config, backUp ...bool) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	if current != nil {
		if current.SyncId != cfg.SyncId {
			return ErrInvalidSyncID
		}
	}
	cfg.SyncId++

	if len(backUp) > 0 && backUp[0] {
		backUpPath := path.Join(Dir, "backup", fmt.Sprintf("%d_config.json", time.Now().UTC().Unix()))
		slog.Info("Backing up config", "sync_id", cfg.SyncId, "to", backUpPath)
		if err := os.Rename(configPath, backUpPath); err != nil {
			slog.Error("Failed to backup config file", "path", backUpPath, "err", err)
		}
	}

	err := write(configPath, cfg)
	if err != nil {
		return err
	}

	*current = *cfg
	return nil
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
