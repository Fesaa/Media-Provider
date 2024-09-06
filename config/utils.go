package config

var current *Config

func OrDefault(value string, defaultValue ...string) string {
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return value
}

func I() *Config {
	return current
}
