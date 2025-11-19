package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/Fesaa/Media-Provider/utils"
)

var (
	// DisableVolumeDirs Kavita has a bug where it rescans series on each scan loop when they have subfolders
	// until this has been resolved, we will not be adding the volume dir
	// See https://github.com/Kareadita/Kavita/issues/3557
	DisableVolumeDirs = boolFeature("DISABLE_VOLUME_DIRS")

	// DisableOneShotInFileName removed the (OneShot) in the filename
	DisableOneShotInFileName = boolFeature("DISABLE_ONE_SHOT_IN_FILE_NAME")

	// SkipTagsOnFailure will skip tag writing if preferences fail to load
	SkipTagsOnFailure = boolFeature("SKIP_TAGS_ON_FAILURE")

	// NoHttpLog skip http altogether, regardless of log levels
	NoHttpLog = boolFeature("NO_HTTP_LOG")
	// ReducedHttpLog log success responses on trace instead of info
	ReducedHttpLog = boolFeature("REDUCED_HTTP_LOG")
	// Development Enable the dev environment
	Development = boolFeature("DEVELOPMENT") || boolFeature("DEV")
	// Docker set when running in docker
	Docker = boolFeature("DOCKER")
	// DatabaseDsn override database dsn. Only takes effect for postgres
	DatabaseDsn = stringFeature("DATABASE_DSN")
	// DbProvider can be postgres or sqlite
	DbProvider = stringFeature("DATABASE_DRIVER", "sqlite")
	// Language set the fallback language
	Language = stringFeature("LANGUAGE")
	// ConfigDir Media-Providers config directory
	ConfigDir = stringFeature("CONFIG_DIR")
	// ConfigFile Media-Providers config fle
	ConfigFile = stringFeature("CONFIG_FILE")
	// TrustedIps passed to fibers trusted ips. Defaults to [10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7, ::1/128]
	TrustedIps = arrayFeature("TRUSTED_IPS", []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7", "::1/128",
	})
	// LimeBaseUrl override the hardcoded lime url
	LimeBaseUrl = stringFeature("LIME_BASE_URL")
	//YtsBaseUrl override the hardcoded yts url
	YtsBaseUrl = stringFeature("YTS_BASE_URL")

	// OtelEndpoint is the endpoint to report traces to. Passed with otlptracehttp.WithEndpointURL
	OtelEndpoint = stringFeature("OTEL_ENDPOINT")
	// OtelAuth is the value for the Authorization header
	OtelAuth = stringFeature("OTEL_AUTH")

	// EnablePprof registers a pprof endpoint to debug issues at runtime
	EnablePprof = boolFeature("ENABLE_PPROF")
)

func arrayFeature[T any](key string, orValue ...[]T) []T {
	val, ok := os.LookupEnv(key)
	if !ok {
		if len(orValue) > 0 {
			return orValue[0]
		}

		return []T{}
	}

	var zero T
	conv := utils.MustReturn(utils.GetConvertor[T](reflect.TypeOf(zero).Kind()))

	parts := strings.Split(val, ",")
	return utils.MaybeMap(parts, func(s string) (T, bool) {
		parsedValue, err := conv(s)
		if err != nil {
			panic(fmt.Errorf("[Feature:%s] error converting %s to T: %w", key, s, err))
		}

		if reflect.ValueOf(parsedValue).IsZero() {
			return zero, false
		}

		return parsedValue, true
	})
}

func boolFeature(key string, orValue ...bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok {
		if len(orValue) > 0 {
			return orValue[0]
		}

		return false
	}

	return strings.ToLower(val) == "true"
}

func intFeature(key string, orValue ...int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		if len(orValue) > 0 {
			return orValue[0]
		}
		return 0
	}

	valInt, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return valInt
}

func stringFeature(key string, orValue ...string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		if len(orValue) > 0 {
			return orValue[0]
		}
		return ""
	}
	return val
}
