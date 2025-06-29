package config

import (
	"github.com/Fesaa/Media-Provider/utils"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var (
	// DisableVolumeDirs Kavita has a bug where it rescans series on each scan loop when they have subfolders
	// until this has been resolved, we will not be adding the volume dir
	// See https://github.com/Kareadita/Kavita/issues/3557
	DisableVolumeDirs = boolFeature("DISABLE_VOLUME_DIRS")

	// DisableOneShotInFileName removed the (OneShot) in the filename
	DisableOneShotInFileName = boolFeature("DISABLE_ONE_SHOT_IN_FILE_NAME")
)

func arrayFeature[T any](key string, f func(string, ...string) T) []T {
	val, ok := envOrValue(key, nil)
	if !ok {
		return []T{}
	}

	parts := strings.Split(val, ",")
	return utils.MaybeMap(parts, func(s string) (T, bool) {
		parsedValue := f(key, s)
		if reflect.ValueOf(parsedValue).IsZero() {
			var zero T
			return zero, false
		}

		return parsedValue, true
	})
}

func boolFeature(key string, orValue ...string) bool {
	val, ok := envOrValue(key, orValue)
	if !ok {
		return false
	}

	return strings.ToLower(val) == "true"
}

func intFeature(key string, orValue ...string) int {
	val, ok := envOrValue(key, orValue)
	if !ok {
		return 0
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return valInt
}

func stringFeature(key string, orValue ...string) string {
	val, ok := envOrValue(key, orValue)
	if !ok {
		return ""
	}
	return val
}

func envOrValue(key string, orValue []string) (string, bool) {
	if len(orValue) > 0 {
		return orValue[0], true
	}
	return os.LookupEnv(key)
}
