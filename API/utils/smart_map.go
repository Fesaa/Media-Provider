package utils

import (
	"strconv"
	"strings"
)

type SmartMap map[string][]string

// SetValue assigns the given values to the key
func (r SmartMap) SetValue(key string, value ...string) {
	r[key] = value
}

// HasKeys returns true if the request includes metadata for all passed keys, false otherwise
func (r SmartMap) HasKeys(keys ...string) bool {
	for _, key := range keys {
		if _, ok := r[key]; !ok {
			return false
		}
	}
	return true
}

// GetStrings returns the values associated with the key as a slice of strings
// An empty slice will return false
func (r SmartMap) GetStrings(key string) ([]string, bool) {
	values, ok := r[key]
	values = Filter(values, func(s string) bool {
		return len(s) > 0
	})
	return values, ok && len(values) > 0
}

// GetString returns the value associated with the key as a string,
// an empty string is returned if not present
func (r SmartMap) GetString(key string, fallback ...string) (string, bool) {
	values, ok := r.GetStrings(key)
	if !ok {
		if len(fallback) > 0 {
			return fallback[0], true
		}
		return "", false
	}

	return values[0], true
}

func (r SmartMap) GetStringOrDefault(key string, fallback string) string {
	s, ok := r.GetString(key)
	if !ok {
		return fallback
	}
	return s
}

// GetInt returns the value associated with the key as an int,
// zero is returned if the value is not present or if conversion failed
func (r SmartMap) GetInt(key string, fallback ...int) (int, error) {
	val, ok := r.GetString(key)
	if !ok {
		if len(fallback) > 0 {
			return fallback[0], nil
		}
		return 0, nil
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		if len(fallback) > 0 {
			return fallback[0], nil
		}
		return 0, err
	}
	return i, nil
}

// GetBool returns the value associated with the key as a bool,
// returns true if the value is equal to "true" while ignoring case
func (r SmartMap) GetBool(key string, fallback ...bool) bool {
	val, ok := r.GetString(key)
	if !ok {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return false
	}
	return strings.ToLower(val) == "true"
}
