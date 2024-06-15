package main

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/providers"
	"log/slog"
	"os"
	"path"
	"strings"
)

func validateConfig() {
	if err := validateRootConfig(config.I()); err != nil {
		slog.Warn("Error validating config: %s", err)
		panic(err)
	}

	for _, p := range config.I().GetPages() {
		if err := validatePage(p); err != nil {
			slog.Warn("Error validating page", "page", p.GetTitle(), "err", err)
			panic(err)
		}
	}

	slog.Info("Config validated")
}

func validateRootConfig(c config.Config) error {
	slog.Debug("Validating root config")
	if strings.HasSuffix(c.GetRootURl(), "/") {
		return fmt.Errorf("invalid root url: %s", c.GetRootURl())
	}
	ok, err := dirExists(c.GetRootDir())
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid root dir: %s", c.GetRootURl())
	}
	return nil
}

func validatePage(page config.Page) error {
	if page.GetTitle() == "" {
		return fmt.Errorf("page title is required")
	}

	for _, p := range page.GetSearchConfig().GetProvider() {
		if !providers.HasProvider(p) {
			return fmt.Errorf("provider %s not found", p)
		}
	}

	ok, err := dirExists(path.Join(config.I().GetRootDir(), page.GetSearchConfig().GetCustomRootDir()))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("customRootDir %s for page %s not found", config.I().GetRootDir(), page.GetTitle())
	}

	for _, dir := range page.GetSearchConfig().GetRootDirs() {
		dir := path.Join(config.I().GetRootDir(), dir)
		ok, err = dirExists(dir)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("rootDir %s for page %s not found", dir, page.GetTitle())
		}
	}

	for name, modifier := range page.GetSearchConfig().GetSearchModifiers() {
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

	if modifier.Type == "" {
		return fmt.Errorf("modifier type is required")
	}

	if !config.IsValidModifierType(modifier.Type) {
		return fmt.Errorf("modifier type '%s' is not a valid. Check the documentation for valid types", modifier.Type)
	}

	for _, pair := range modifier.Values {
		if pair.Name == "" {
			return fmt.Errorf("modifier value name is required")
		}
		if pair.Key == "" {
			return fmt.Errorf("modifier value key is required")
		}
	}

	return nil
}

func dirExists(path string) (bool, error) {
	slog.Debug("Checking directory", "path", path)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
