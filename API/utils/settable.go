package utils

// Settable represents a variable where the zero value has more meaning than undecided yet
// A Settable may be Set more than once
type Settable[T any] struct {
	val T
	set bool
}

// Set sets the value
func (s *Settable[T]) Set(t T) {
	s.set = true
	s.val = t
}

// Get returns the value if present, otherwise zero, false
func (s *Settable[T]) Get() (T, bool) {
	if !s.set {
		var zero T
		return zero, false
	}
	return s.val, true
}
