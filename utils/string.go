package utils

import (
	"regexp"
	"strconv"
	"strings"
)

var re = regexp.MustCompile(`[^a-zA-Z0-9]`)

func OrElse(s string, def string) string {
	if len(s) == 0 {
		return def
	}
	return s
}

func NonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func Normalize(s string) string {
	return strings.ToLower(re.ReplaceAllString(s, ""))
}

// PadInt returns the i as a string s, with 0's added to the left until len(s) >= n
func PadInt(i int, n int) string {
	return pad(strconv.Itoa(i), n)
}

// PadFloatFromString returns the float as a string, with pad called on the whole part
// and the decimal part copied from the input string, if present
// This method assumes the string is a valid float
func PadFloatFromString(s string, n int) string {
	parts := strings.Split(s, ".")
	if len(parts) < 2 { // No decimal part
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
