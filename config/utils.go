package config

func OrDefault(value string, defaultValue ...string) string {
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return value
}
