package utils

import "sync"

type SafeMap[K comparable, V any] interface {
	Has(k K) bool
	Get(k K) (V, bool)
	Set(k K, v V)
	Len() int
	Delete(k K)
	Clear()
	Keys() []K
	Values() []V
	ForEach(f func(k K, v V))
	ForEachSafe(f func(k K, v V))
	Any(f func(k K, v V) bool) bool
	Count(f func(k K, v V) bool) int
	Find(f func(k K, v V) bool) (*V, bool)
}

type safeMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

func NewSafeMap[K comparable, V any](m ...map[K]V) SafeMap[K, V] {
	var startMap map[K]V
	if len(m) > 0 {
		startMap = m[0]
	} else {
		startMap = make(map[K]V)
	}
	return &safeMap[K, V]{
		m:    startMap,
		lock: sync.RWMutex{},
	}
}

func (s *safeMap[K, V]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m = make(map[K]V)
}

func (s *safeMap[K, V]) Count(f func(K, V) bool) int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	i := 0
	for k, v := range s.m {
		if f(k, v) {
			i++
		}
	}
	return i
}

func (s *safeMap[K, V]) Find(f func(k K, v V) bool) (*V, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.m {
		if f(k, v) {
			return &v, true
		}
	}
	return nil, false
}

func (s *safeMap[K, V]) Any(f func(k K, v V) bool) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.m {
		if f(k, v) {
			return true
		}
	}
	return false
}

func (s *safeMap[K, V]) Has(k K) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.m[k]
	return ok
}

func (s *safeMap[K, V]) Get(k K) (V, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.m[k]
	return v, ok
}

func (s *safeMap[K, V]) Set(k K, v V) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m[k] = v
}

func (s *safeMap[K, V]) Delete(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.m, k)
}

func (s *safeMap[K, V]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.m)
}

// ForEachSafe iterates over the map, does not unlock during function all.
//
// Use SafeMap.ForEach if you need to modify the map
func (s *safeMap[K, V]) ForEachSafe(f func(K, V)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.m {
		f(k, v)
	}
}

// ForEach iterates over the map, and unlocks during function call
func (s *safeMap[K, V]) ForEach(f func(K, V)) {
	s.lock.Lock()
	for k, v := range s.m {
		s.lock.Unlock()
		f(k, v)
		s.lock.Lock()
	}
	s.lock.Unlock()
}

func (s *safeMap[K, V]) Keys() []K {
	s.lock.RLock()
	defer s.lock.RUnlock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

func (s *safeMap[K, V]) Values() []V {
	s.lock.RLock()
	defer s.lock.RUnlock()
	values := make([]V, 0, len(s.m))
	for _, v := range s.m {
		values = append(values, v)
	}
	return values
}
