package utils

import (
	"errors"
	"time"
)

var (
	ErrCachedItemExpired = errors.New("cached item has expired")
)

// CachedItem is the simplest way of caching an item, do not use for more complicated stuff
// does not evict anything itself
type CachedItem[T any] interface {
	Get() (T, error)
	HasExpired() bool
}

type cachedItem[T any] struct {
	data      T
	expiredAt time.Time
}

func (c *cachedItem[T]) Get() (T, error) {
	if c.HasExpired() {
		var zero T
		return zero, ErrCachedItemExpired
	}

	return c.data, nil
}

func (c *cachedItem[T]) HasExpired() bool {
	return c.expiredAt.Before(time.Now())
}

func NewCachedItem[T any](data T, exp time.Duration) CachedItem[T] {
	return &cachedItem[T]{
		data:      data,
		expiredAt: time.Now().Add(exp),
	}
}
