package utils

import "sync"

type SafeMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

func NewSafeMap[K comparable, V any](m ...map[K]V) *SafeMap[K, V] {
	var startMap map[K]V
	if len(m) > 0 {
		startMap = m[0]
	} else {
		startMap = make(map[K]V)
	}
	return &SafeMap[K, V]{
		m:    startMap,
		lock: sync.RWMutex{},
	}
}

func (s *SafeMap[K, V]) Has(k K) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, ok := s.m[k]
	return ok
}

func (s *SafeMap[K, V]) Get(k K) (V, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	v, ok := s.m[k]
	return v, ok
}

func (s *SafeMap[K, V]) Set(k K, v V) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m[k] = v
}

func (s *SafeMap[K, V]) Delete(k K) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.m, k)
}

func (s *SafeMap[K, V]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.m)
}

// ForEachSafe iterates over the map, does not unlock during functionc all.
//
// Use SafeMap.ForEach if you need to modify the map
func (s *SafeMap[K, V]) ForEachSafe(f func(K, V)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.m {
		f(k, v)
	}
}

// ForEach iterates over the map, and unlocks during function call
func (s *SafeMap[K, V]) ForEach(f func(K, V)) {
	s.lock.Lock()
	for k, v := range s.m {
		s.lock.Unlock()
		f(k, v)
		s.lock.Lock()
	}
	s.lock.Unlock()
}
