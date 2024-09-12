package main

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/go-playground/validator/v10"
	"os"
	"path"
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

	for _, p := range cfg.Pages {
		if err := validatePage(p); err != nil {
			log.Warn("error while validating page", "page", p.Title, "err", err)
			panic(err)
		}
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

	if changed {
		log.Warn("BaseUrl was forcefully changed, saving config", "baseUrl", c.BaseUrl)
		return c.Save()
	}

	return nil
}

func validatePage(page config.Page) error {
	if page.Title == "" {
		return fmt.Errorf("page title is required")
	}

	for _, p := range page.Provider {
		if !providers.HasProvider(p) {
			return fmt.Errorf("provider %s not found", p)
		}
	}

	rootPath := path.Join(cfg.GetRootDir(), page.CustomRootDir)
	ok, err := dirExists(rootPath)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("customRootDir does not exist %s", rootPath)
	}

	for _, dir := range page.Dirs {
		dir := path.Join(cfg.GetRootDir(), dir)
		ok, err = dirExists(dir)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("rootDir %s for page %s not found", dir, page.Title)
		}
	}

	for name, modifier := range page.Modifiers {
		if err = validateModifier(modifier); err != nil {
			return fmt.Errorf("invalid search modifier '%s': %s", name, err)
		}
	}

	return nil
}

func validateModifier(modifier config.Modifier) error {
	if modifier.Title == "" {
		return fmt.Errorf("modifier title is required")
	}

	if modifier.Type == 0 {
		return fmt.Errorf("modifier type is required")
	}

	if !config.IsValidModifierType(modifier.Type) {
		return fmt.Errorf("modifier type '%s' is not a valid. Check the documentation for valid types", modifier.Type)
	}

	for name, key := range modifier.Values {
		if name == "" {
			return fmt.Errorf("modifier value name is required")
		}
		if key == "" {
			return fmt.Errorf("modifier value key is required")
		}
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
