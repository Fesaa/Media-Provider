package utils

import "errors"

// Settable represents a variable where the zero value has more meaning than undecided yet
// A Settable may be Set more than once
type Settable[T any] struct {
	val T
	set bool
}

var ErrNotSet = errors.New("settable has not been set yet")

// Set sets the value
func (s *Settable[T]) Set(t T) {
	s.set = true
	s.val = t
}

// Get returns the value or ErrNotSet if not set
func (s *Settable[T]) Get() (T, error) {
	if !s.set {
		var zero T
		return zero, ErrNotSet
	}
	return s.val, nil
}
