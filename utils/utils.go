package utils

import (
	"os"
	"strings"
)

func GetEnv(key string, defaultVal ...string) string {
	val, ok := os.LookupEnv(key)
	if !ok && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return val
}

func GetBoolEnv(key string, defaultVal ...bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	val = strings.ToLower(val)
	return val == "true" || val == "t" || val == "1"
}

func FeatureEnabled(key string) bool {
	return GetBoolEnv(key, true)
}
