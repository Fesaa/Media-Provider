package main

import (
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

func validateConfig(cfg *config.Config, log zerolog.Logger, val *validator.Validate, fs afero.Afero) error {
	log = log.With().Str("handler", "config-validator").Logger()

	if err := val.Struct(cfg); err != nil {
		return err
	}

	if err := validateRootConfig(cfg, log); err != nil {
		return err
	}

	log.Info().Msg("Config validated")
	return nil
}

func validateRootConfig(c *config.Config, log zerolog.Logger) error {
	log.Debug().Msg("Validating root config")

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
