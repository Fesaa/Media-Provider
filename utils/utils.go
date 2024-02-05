package utils

import "os"

func GetEnv(key string, defaultVal ...string) string {
	val, ok := os.LookupEnv(key)
	if !ok && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return val
}
