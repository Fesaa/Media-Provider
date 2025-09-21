package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

// WaitFor waits for at most d, and at least until wg has finished
func WaitFor(wg *sync.WaitGroup, d time.Duration) {
	select {
	case <-Wait(wg):
		return
	case <-time.After(d):
		return
	}
}

// Wait returns a channel which completes when the WaitGroup has finished
func Wait(wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

// Defer adds 1 to the WaitGroup, then starts a goroutine called f and deferring wg.Done
func Defer(f func() error, log zerolog.Logger, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := f(); err != nil {
			log.Error().Err(err).Msg("Deferred function failed")
		}
	}()
}

// Clamp returns at least minV, and at most maxV
//
// Will panic if maxV < minV or minV > maxV
func Clamp(val, minV, maxV int) int {
	if (minV > maxV) || (maxV < minV) {
		panic(fmt.Sprintf("Clamp(?, %v, %v) out of range", val, minV))
	}

	return max(minV, min(maxV, val))
}

func GenerateSecret(length int) (string, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(secret), nil
}

func GenerateUrlSecret(length int) (string, error) {
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

type Invoker interface {
	Invoke(function interface{}, opts ...dig.InvokeOption) (err error)
}

// MustInvoke tries construction T with the given Container
// Ensure the needed method has been Provider to the Container
func MustInvoke[T any](c Invoker) T {
	var t T
	Must(c.Invoke(func(myT T) {
		t = myT
	}))
	return t
}

func Must(err error) {
	if err != nil {
		panic(PrettyErrorString(err))
	}
}

func MustReturn[T any](result T, err error) T {
	if err != nil {
		panic(PrettyErrorString(err))
	}
	return result
}

func MustReturn2[T any](_ any, result T, err error) T {
	if err != nil {
		panic(PrettyErrorString(err))
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

func IdentityFunc[T any]() func(T) T {
	return func(t T) T {
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

func Ternary[T any](condition bool, ifTrue, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
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

func PrettyErrorString(err error) string {
	if err == nil {
		return "No error."
	}

	var b strings.Builder
	b.WriteString("Error trace:\n")
	for i := 1; err != nil; i++ {
		b.WriteString(fmt.Sprintf("  [%d] %v\n", i, err))
		err = errors.Unwrap(err)
	}
	return b.String()
}
