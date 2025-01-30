package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
)

func GenerateSecret(length int) (string, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(secret), nil
}

func GenerateApiKey() (string, error) {
	bytes := make([]byte, 16)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	apiKey := hex.EncodeToString(bytes)
	return apiKey, nil
}

func HumanReadableSpeed(s int64) string {
	speed := float64(s)
	if speed < 1024 {
		return fmt.Sprintf("%.2f B/s", speed)
	}
	speed /= 1024
	if speed < 1024 {
		return fmt.Sprintf("%.2f KB/s", speed)
	}
	speed /= 1024
	return fmt.Sprintf("%.2f MB/s", speed)
}

// Percent returns the % of a contained in b. At most 100 is returned. Percent(a, 0) = 100 for a != 0, Percent(0,0) = 0
func Percent(a, b int64) int64 {
	if b == 0 {
		return (int64)(Ternary(a == 0, 0, 100))
	}
	b = max(b, a)
	ratio := (float64)(a) / (float64)(b)
	return (int64)(ratio * 100)
}

var sizes = [...]string{"Bytes", "KB", "MB", "GB", "TB"}

func BytesToSize(bytes float64) string {
	if bytes == 0 {
		return "0 Byte"
	}
	i := math.Floor(math.Log(bytes) / math.Log(1024))
	return fmt.Sprintf("%.2f %s", bytes/math.Pow(1024, i), sizes[int(i)])
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustReturn[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}

func MustHave[T any](result T, ok bool) T {
	if !ok {
		panic("MustHave[T] was not true")
	}
	return result
}

func Identity[T any](t T) func() T {
	return func() T {
		return t
	}
}

func Stringify(i int) string {
	return fmt.Sprintf("%d", i)
}

func OrDefault[T any](array []T, defaultValue T) T {
	if len(array) == 0 {
		return defaultValue
	}
	return array[0]
}

func Ternary[T any](condition bool, tuple ...T) T {
	if len(tuple) != 2 {
		panic("Tuple should be of length 2")
	}
	if condition {
		return tuple[0]
	}
	return tuple[1]
}

func SortFloats(a, b string) int {
	if a == b {
		return 0
	}

	if a == "" {
		return 1
	}

	if b == "" {
		return -1
	}

	fa, _ := strconv.ParseFloat(a, 64)
	fb, _ := strconv.ParseFloat(b, 64)
	return (int)(fb - fa)
}
