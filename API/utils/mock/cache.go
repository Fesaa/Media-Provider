package mock

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/config"
)

type Cache struct{}

func (c Cache) GetWithContext(ctx context.Context, s string) ([]byte, error) {
	return nil, nil
}

func (c Cache) SetWithContext(ctx context.Context, s string, bytes []byte, duration time.Duration) error {
	return nil
}

func (c Cache) DeleteWithContext(ctx context.Context, s string) error {
	return nil
}

func (c Cache) ResetWithContext(ctx context.Context) error {
	return nil
}

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

func (c Cache) DefaultExpiration() time.Duration {
	return 0
}
