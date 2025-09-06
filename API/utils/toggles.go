package utils

type Toggles[K comparable] struct {
	m SafeMap[K, bool]
}

func NewToggles[K comparable]() *Toggles[K] {
	return &Toggles[K]{
		m: NewSafeMap[K, bool](),
	}
}

func (t *Toggles[K]) Toggle(k K) {
	cur, ok := t.m.Get(k)

	t.m.Set(k, !(cur && ok))
}

func (t *Toggles[K]) Toggled(k K) bool {
	cur, ok := t.m.Get(k)
	return ok && cur
}
