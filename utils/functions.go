package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go.uber.org/dig"
	"math"
	"strconv"
	"strings"
	"time"
)

// Clamp returns at least minV, and at most maxV
//
// Will panic if maxV < minV or minV > maxV
func Clamp(val, minV, maxV int) int {
	if (minV > maxV) || (maxV < minV) {
		panic(fmt.Sprintf("Clamp(?, %v, %v) out of range", val, minV))
	}

	return max(minV, min(maxV, val))
}

// TryCatch will return the consumed value if the error returned by the producer was nil. Other returns the fallback
// value and calls the first errorHandler if present
func TryCatch[T, U any](
	producer func() (T, error),
	mapper func(T) U,
	fallback U,
	errorHandlers ...func(error),
) U {
	t, err := producer()
	if err != nil {
		if len(errorHandlers) > 0 {
			errorHandlers[0](err)
		}
		return fallback
	}

	return mapper(t)
}

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

// MustInvoke tries construction T with the given scope
// Ensure the needed method has been Provider to the scope
func MustInvoke[T any](c *dig.Scope) T {
	var t T
	Must(c.Invoke(func(myT T) {
		t = myT
	}))
	return t
}

// MustInvokeCont tries construction T with the given Container
// Ensure the needed method has been Provider to the Container
func MustInvokeCont[T any](c *dig.Container) T {
	var t T
	Must(c.Invoke(func(myT T) {
		t = myT
	}))
	return t
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

func Shorten(s string, length int) string {
	if len(s) < length {
		return s
	}

	s = strings.SplitN(s, "\n", 1)[0]
	if len(s) < 4 {
		return s[:length]
	}

	return s[:length-3] + "..."
}

func IsSameDay(t1, t2 time.Time) bool {
	d1, m1, y1 := t1.Date()
	d2, m2, y2 := t2.Date()

	return d1 == d2 && m1 == m2 && y1 == y2
}

func Count[T any](array []T, f func(T) bool) int {
	i := 0
	for _, t := range array {
		if f(t) {
			i++
		}
	}
	return i
}
