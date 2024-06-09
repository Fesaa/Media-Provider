package utils

import "sync"

type SafeMap[K comparable, V any] struct {
	lock sync.RWMutex
	m    map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m:    make(map[K]V),
		lock: sync.RWMutex{},
	}
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

func (s *SafeMap[K, V]) ForEachSafe(f func(K, V)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for k, v := range s.m {
		f(k, v)
	}
}

func (s *SafeMap[K, V]) ForEach(f func(K, V)) {
	s.lock.Lock()
	for k, v := range s.m {
		s.lock.Unlock()
		f(k, v)
		s.lock.Lock()
	}
	s.lock.Unlock()
}

func (s *SafeMap[K, V]) Lock() {
	s.lock.Lock()
}

func (s *SafeMap[K, V]) Unlock() {
	s.lock.Unlock()
}
