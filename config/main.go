package config

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/providers"
	"os"

	"gopkg.in/yaml.v3"
)

var c configImpl

func I() Config {
	return c
}

func OrDefault(value string, defaultValue ...string) string {
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return value
}

func validate() error {
	for i, page := range c.Pages {
		if !providers.HasProvider(page.SearchConfig.Provider) {
			return fmt.Errorf("page %d (%s) has an invalid provider '%s'", i, page.Title, page.SearchConfig.Provider)
		}
	}

	return nil
}

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return err
	}

	return validate()
}
