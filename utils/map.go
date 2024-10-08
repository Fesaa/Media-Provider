package utils

func Map[T, S any](in []T, f func(T) S) []S {
	out := make([]S, len(in))
	for i, t := range in {
		out[i] = f(t)
	}
	return out
}

func MaybeMap[T, S any](in []T, f func(T) (S, bool)) []S {
	out := make([]S, 0)
	for _, t := range in {
		if s, ok := f(t); ok {
			out = append(out, s)
		}
	}
	return out
}

func Filter[T any](in []T, f func(T) bool) []T {
	out := make([]T, 0)
	for _, t := range in {
		if f(t) {
			out = append(out, t)
		}
	}
	return out
}

func Find[T any](in []T, f func(T) bool) *T {
	for _, t := range in {
		if f(t) {
			return &t
		}
	}

	return nil
}

func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func Values[K comparable, V any](m map[K]V) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
