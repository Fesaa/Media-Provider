package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func PadInt(i int, n int) string {
	return pad(strconv.Itoa(i), n)
}

func PadFloat(f float64, n int) string {
	full := fmt.Sprintf("%.1f", f)
	parts := strings.Split(full, ".")
	if len(parts) < 2 || parts[1] == "0" { // No decimal part
		return pad(parts[0], n)
	}
	return pad(parts[0], n) + "." + parts[1]
}

func pad(str string, n int) string {
	if len(str) < n {
		str = strings.Repeat("0", n-len(str)) + str
	}
	return str
}
