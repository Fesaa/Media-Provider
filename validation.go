package main

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

func validateConfig(cfg *config.Config, log zerolog.Logger, val *validator.Validate) error {
	log = log.With().Str("handler", "config-validator").Logger()

	if err := val.Struct(cfg); err != nil {
		return err
	}

	if err := validateRootConfig(cfg, log); err != nil {
		return err
	}

	if cfg.Downloader.MaxConcurrentTorrents < 1 || cfg.Downloader.MaxConcurrentTorrents > 10 {
		return fmt.Errorf("max concurrent torrents must be between 1 and 10")
	}

	if cfg.Downloader.MaxConcurrentMangadexImages < 1 || cfg.Downloader.MaxConcurrentMangadexImages > 5 {
		return fmt.Errorf("max concurrent mangadex images must be between 1 and 5")
	}

	log.Info().Msg("Config validated")
	return nil
}

func validateRootConfig(c *config.Config, log zerolog.Logger) error {
	log.Debug().Msg("Validating root config")
	if strings.HasSuffix(c.GetRootDir(), "/") {
		return fmt.Errorf("invalid root url, must not end with /: %s", c.GetRootDir())
	}
	ok, err := dirExists(c.GetRootDir())
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid root dir, does not exist: %s", c.GetRootDir())
	}

	// User can easily forget to add a / to the base url, so we add it for them
	// The meaning of it doesn't change, as what would happen if we made directories for them in the above;
	changed := false
	if !strings.HasPrefix(c.BaseUrl, "/") {
		changed = true
		c.BaseUrl = "/" + c.BaseUrl
	}

	if !strings.HasSuffix(c.BaseUrl, "/") {
		changed = true
		c.BaseUrl += "/"
	}

	if c.Secret == "" {
		secret, err := utils.GenerateSecret(64)
		if err != nil {
			return err
		}
		c.Secret = secret
		changed = true
	}

	if changed {
		log.Warn().Msg("Config was changed by validateRootConfig, saving...")
		return c.Save(c)
	}

	return nil
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
