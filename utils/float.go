package utils

import "strconv"

// SafeFloat parses the string as a float64, and returns -1 on error
func SafeFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return f
}
