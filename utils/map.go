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
