package utils

import "slices"

// AtIdx returns the value at the given index, unless out of range; then return the zero value for the element
func AtIdx[T any](s []T, index int) T {
	if index >= len(s) || index < 0 {
		var zero T
		return zero
	}

	return s[index]
}

func ForEach[T any](s []T, f func(T)) {
	for _, t := range s {
		f(t)
	}
}

func Contains[E comparable](s []E, s2 []E) bool {
	for _, t := range s {
		if slices.Contains(s2, t) {
			return true
		}
	}
	return false
}
