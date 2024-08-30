package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
)

func Load(path string) (*Config, error) {
	cfg, err := Read(path)

	if errors.Is(err, os.ErrNotExist) {
		slog.Warn("Config file not found, creating new one", "path", path)
		cfg = defaultConfig()
		err = Write(path, cfg)
	}

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func Write(path string, cfg *Config) error {
	slog.Debug("Writing config", "path", path)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
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
