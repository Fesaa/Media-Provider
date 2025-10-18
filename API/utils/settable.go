package utils

// Settable represents a variable where the zero value has more meaning than undecided yet
// A Settable may be Set more than once
type Settable[T any] struct {
	val T
	set bool
}

// NewSettable returns a Settable with the first passed value. If no value is passed, returns a non set Settable
func NewSettable[T any](v ...T) Settable[T] {
	if len(v) == 0 {
		return Settable[T]{}
	}

	return Settable[T]{val: v[0], set: true}
}

// NewSettableFromErr returns a set Settable only if err is nil
func NewSettableFromErr[T any](value T, err error) Settable[T] {
	if err != nil {
		return Settable[T]{}
	}

	return Settable[T]{val: value, set: true}
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
