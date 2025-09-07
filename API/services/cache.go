package services

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type CacheService interface {
	fiber.Storage
	Type() config.CacheType
}

type cacheService struct {
	fiber.Storage
	log       zerolog.Logger
	cacheType config.CacheType
}

func CacheServiceProvider(log zerolog.Logger, service SettingsService) (CacheService, error) {
	cs := &cacheService{
		log: log.With().Str("handler", "cache-service").Logger(),
	}

	settings, err := service.GetSettingsDto(context.Background())
	if err != nil {
		return nil, err
	}

	switch settings.CacheType {
	case config.REDIS:
		cs.Storage = utils.NewRedisCacheStorage(log, "cache-service", settings.RedisAddr)
	case config.MEMORY:
		cs.Storage = newMemoryCache()
	default:
		cs.log.Fatal().Any("config", settings.CacheType).Msg("invalid cache type")
	}

	return cs, nil
}

func (s *cacheService) Type() config.CacheType {
	return s.cacheType
}

type item struct {
	value []byte
	exp   time.Time
}

type memoryCache struct {
	store utils.SafeMap[string, item]
}

func (m *memoryCache) Get(key string) ([]byte, error) {
	val, ok := m.store.Get(key)
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val.value, nil
}

func (m *memoryCache) Set(key string, val []byte, exp time.Duration) error {
	m.store.Set(key, item{val, time.Now().Add(exp)})
	return nil
}

func (m *memoryCache) Delete(key string) error {
	m.store.Delete(key)
	return nil
}

func (m *memoryCache) Reset() error {
	m.store.Clear()
	return nil
}

func (m *memoryCache) Close() error {
	return nil
}

func (m *memoryCache) gc() {
	for range time.Tick(time.Second) {
		if m.store.Len() == 0 {
			continue
		}

		m.store.ForEach(func(k string, v item) {
			if time.Now().After(v.exp) {
				m.store.Delete(k)
			}
		})
	}
}

func newMemoryCache() *memoryCache {
	mc := &memoryCache{
		store: utils.NewSafeMap[string, item](),
	}
	go mc.gc()
	return mc
}
