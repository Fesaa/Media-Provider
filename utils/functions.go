package utils

import (
	"fmt"
	"math"
)

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

func Percent(a, b int64) int64 {
	b = max(b, 1)
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
