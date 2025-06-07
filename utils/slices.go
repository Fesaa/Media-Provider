package utils

// AtIdx returns the value at the given index, unless at of range; then return the zero value for the element
func AtIdx[T any](s []T, index int) T {
	if index >= len(s) || index < 0 {
		var zero T
		return zero
	}

	return s[index]
}
