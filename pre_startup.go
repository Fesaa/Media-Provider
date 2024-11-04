package main

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-playground/validator/v10"
	"os"
	"strings"
)

func validateConfig(cfg *config.Config) {
	if err := validator.New().Struct(cfg); err != nil {
		log.Warn("error while validating config", "err", err)
		panic(err)
	}

	if err := validateRootConfig(cfg); err != nil {
		log.Warn("error while validating config", "err", err)
		panic(err)
	}

	if cfg.Downloader.MaxConcurrentTorrents < 1 || cfg.Downloader.MaxConcurrentTorrents > 10 {
		log.Warn("invalid max concurrent torrents", "value", cfg.Downloader.MaxConcurrentTorrents)
		panic("invalid max concurrent torrents")
	}

	if cfg.Downloader.MaxConcurrentMangadexImages < 1 || cfg.Downloader.MaxConcurrentMangadexImages > 5 {
		log.Warn("invalid max concurrent mangadex images", "value", cfg.Downloader.MaxConcurrentMangadexImages)
		panic("invalid max concurrent mangadex images")
	}

	log.Info("Config validated")
}

func validateRootConfig(c *config.Config) error {
	log.Debug("Validating root config")
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
		log.Warn("Config was changed by validateRootConfig, saving...")
		return c.Save()
	}

	return nil
}

func dirExists(path string) (bool, error) {
	log.Trace("checking directory", "path", path)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
