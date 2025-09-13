package utils

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// GetConvertor returns a method that converse a string to the desired type. Not all types are supported, check
// implementation details
func GetConvertor[T any]() (func(string) (T, error), error) {
	var zero T

	switch any(zero).(type) {
	case int:
		return func(s string) (T, error) {
			val, err := strconv.Atoi(s)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case int64:
		return func(s string) (T, error) {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case int32:
		return func(s string) (T, error) {
			val, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return zero, err
			}
			return any(int32(val)).(T), nil
		}, nil
	case uint:
		return func(s string) (T, error) {
			val, err := strconv.ParseUint(s, 10, 0)
			if err != nil {
				return zero, err
			}
			return any(int(val)).(T), nil
		}, nil
	case uint64:
		return func(s string) (T, error) {
			val, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case float32:
		return func(s string) (T, error) {
			val, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return zero, err
			}
			return any(float32(val)).(T), nil
		}, nil
	case float64:
		return func(s string) (T, error) {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case bool:
		return func(s string) (T, error) {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case string:
		return func(s string) (T, error) {
			return any(s).(T), nil
		}, nil
	case time.Time:
		return func(s string) (T, error) {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return zero, err
			}
			return any(t).(T), nil
		}, nil
	case net.IP:
		return func(s string) (T, error) {
			ip := net.ParseIP(s)
			if ip == nil {
				return zero, &net.ParseError{Type: "IP address", Text: s}
			}
			return any(ip).(T), nil
		}, nil
	case net.IPNet:
		return func(s string) (T, error) {
			_, cidr, err := net.ParseCIDR(s)
			if err != nil {
				return zero, err
			}

			return any(cidr).(T), nil
		}, nil
	}

	return nil, fmt.Errorf("unknown type %T", any(zero))
}
