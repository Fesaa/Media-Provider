package utils

import (
	"sync"
	"time"
)

type Cache[T any] struct {
	objects map[string]*CacheObject[T]
	lock    *sync.RWMutex
	expiry  time.Duration
}

func NewCache[T any](expiry time.Duration) *Cache[T] {
	c := &Cache[T]{
		objects: make(map[string]*CacheObject[T]),
		lock:    &sync.RWMutex{},
		expiry:  expiry,
	}
	go c.cleaner()
	return c
}

func (c *Cache[T]) Get(key string) *T {
	c.lock.RLock()
	defer c.lock.RUnlock()
	obj, ok := c.objects[key]
	if !ok {
		return nil
	}
	return &obj.Obj
}

func (c *Cache[T]) Set(key string, obj T) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.objects[key] = newCacheObject(obj, c.expiry)
}

func (c *Cache[T]) cleaner() {
	for range time.Tick(c.expiry) {
		c.lock.Lock()
		for k, v := range c.objects {
			if v.Expired() {
				delete(c.objects, k)
			}
		}
		c.lock.Unlock()
	}
}

type CacheObject[T any] struct {
	Obj    T
	insert time.Time
	expiry time.Duration
}

func newCacheObject[T any](obj T, expiry time.Duration) *CacheObject[T] {
	return &CacheObject[T]{
		Obj:    obj,
		insert: time.Now(),
		expiry: expiry,
	}
}

func (c *CacheObject[T]) Expired() bool {
	return time.Since(c.insert) > c.expiry
}
