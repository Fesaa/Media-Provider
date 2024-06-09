package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

var C Config

func I() Config {
	return C
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

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &C)
}

func SetPages(pages []Page) {
	C.Pages = pages
}

func ReloadPages(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return err
	}

	C.Pages = c.Pages
	return nil
}
