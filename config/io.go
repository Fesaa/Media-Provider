package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path"
)

var configPath string

func init() {
	file := OrDefault(os.Getenv("CONFIG_FILE"), "config.json")
	configPath = path.Join("", file)
}

func Load() (*Config, error) {
	cfg, err := Read(configPath)

	if errors.Is(err, os.ErrNotExist) {
		slog.Warn("Config file not found, creating new one", "path", configPath)
		cfg = defaultConfig()
		err = Write(configPath, cfg)
	}

	if err != nil {
		return nil, err
	}

	current = cfg
	return cfg, nil
}

func Save(cfg *Config) error {
	return Write(configPath, cfg)
}

func (c *Config) Save() error {
	return Save(c)
}

func Write(path string, cfg *Config) error {
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

func Read(path string) (*Config, error) {
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
