package utils

func MapWithIdx[T, S any](in []T, f func(int, T) S) []S {
	out := make([]S, len(in))
	for i, t := range in {
		out[i] = f(i, t)
	}
	return out
}

func Map[T, S any](in []T, f func(T) S) []S {
	return MapWithIdx(in, func(_ int, t T) S {
		return f(t)
	})
}

func FlatMapMany[E any](in ...[]E) []E {
	var out []E
	for _, e := range in {
		out = append(out, e...)
	}
	return out
}

func FlatMap[E any](in [][]E) []E {
	var out []E
	for _, e := range in {
		out = append(out, e...)
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

// Find returns the first element in the slice that returns true for the function
func Find[T any](in []T, f func(T) bool) *T {
	for _, t := range in {
		if f(t) {
			return &t
		}
	}

	return nil
}

func Any[T any](in []T, f func(T) bool) bool {
	return Find(in, f) != nil
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

func MapKeys[K comparable, V, T any](m map[K]V, f func(K) T) []T {
	out := make([]T, 0, len(m))
	for k := range m {
		out = append(out, f(k))
	}
	return out
}

func MapValues[K comparable, V, T any](m map[K]V, f func(V) T) []T {
	out := make([]T, 0, len(m))
	for _, v := range m {
		out = append(out, f(v))
	}
	return out
}

func GroupBy[K comparable, V any](s []V, f func(V) K) map[K][]V {
	out := map[K][]V{}

	for _, v := range s {
		key := f(v)
		out[key] = append(out[key], v)
	}

	return out
}

func MapToString[T ~string](m []T) []string {
	return Map(m, func(t T) string {
		return string(t)
	})
}
