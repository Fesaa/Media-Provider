package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`[^a-zA-Z0-9]`)

func Normalize(s string) string {
	return strings.ToLower(re.ReplaceAllString(s, ""))
}

// PadInt returns the i as a string s, with 0's added to the left until len(s) >= n
func PadInt(i int, n int) string {
	return pad(strconv.Itoa(i), n)
}

// PadFloat returns the float as a string, with pad called on the whole part
// and the first decimal part added, if present
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
