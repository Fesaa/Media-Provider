package mock

import (
	"github.com/Fesaa/Media-Provider/config"
	"time"
)

type Cache struct{}

func (c Cache) Get(key string) ([]byte, error) {
	return nil, nil
}

func (c Cache) Set(key string, val []byte, exp time.Duration) error {
	return nil
}

func (c Cache) Delete(key string) error {
	return nil
}

func (c Cache) Reset() error {
	return nil
}

func (c Cache) Close() error {
	return nil
}

func (c Cache) Type() config.CacheType {
	return config.MEMORY
}
