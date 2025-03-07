package services

import (
	"context"
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"time"
)

type CacheService interface {
	fiber.Storage
	Type() config.CacheType
}

type cacheService struct {
	log zerolog.Logger
	cfg *config.Config

	fiber.Storage
}

func CacheServiceProvider(cfg *config.Config, log zerolog.Logger) CacheService {
	cs := &cacheService{
		log: log.With().Str("handler", "cache-service").Logger(),
		cfg: cfg,
	}
	cs.constructCache()
	return cs
}

func (s *cacheService) constructCache() {
	if s.cfg.Cache.Type == config.REDIS {
		s.log.Debug().Msg("trying to use redis cache")
		s.Storage = newRedisCacheStorage(s.cfg, s.log)
	}
	if s.Storage == nil {
		s.log.Debug().Msg("Falling back to in-memory cache")
		c := &memoryStorage{utils.NewSafeMap[string, *cacheEntry](), s.log}
		go c.cleaner()
		s.Storage = c
	}
}

func newRedisCacheStorage(cfg *config.Config, log zerolog.Logger) fiber.Storage {
	rds := redis.NewClient(&redis.Options{
		Addr:           cfg.Cache.RedisAddr,
		Password:       "",
		DB:             0,
		ClientName:     "cache-service",
		IdentitySuffix: "media-provider",
	})

	if err := rds.Ping(context.Background()).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to connect to redis")
		return nil
	}

	return &redisWrapper{rdb: rds, log: log}
}

func (s *cacheService) Type() config.CacheType {
	return s.cfg.Cache.Type
}

type redisWrapper struct {
	rdb *redis.Client
	log zerolog.Logger
}

func (r *redisWrapper) Get(key string) ([]byte, error) {
	b, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		r.log.Trace().Err(err).Str("key", key).Msg("redis returned an error")
	}
	return b, err
}

func (r *redisWrapper) Set(key string, val []byte, exp time.Duration) error {
	err := r.rdb.Set(context.Background(), key, val, exp).Err()
	if err != nil {
		r.log.Trace().Err(err).Str("key", key).Msg("redis returned an error")
	}
	return err
}

func (r *redisWrapper) Delete(key string) error {
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *redisWrapper) Reset() error {
	return r.rdb.FlushAll(context.Background()).Err()
}

func (r *redisWrapper) Close() error {
	return r.rdb.Close()
}

type cacheEntry struct {
	bytes []byte
	exp   time.Time
}

func (c *cacheEntry) Expired() bool {
	return c.exp.Before(time.Now())
}

type memoryStorage struct {
	_map utils.SafeMap[string, *cacheEntry]
	log  zerolog.Logger
}

func (s *memoryStorage) cleaner() {
	for range time.Tick(5 * time.Second) {
		if s._map.Len() == 0 {
			continue
		}
		s._map.ForEach(func(k string, v *cacheEntry) {
			if v.Expired() {
				s.log.Trace().Str("key", k).Msg("cache expired")
				s._map.Delete(k)
			}
		})
	}
}

func (s *memoryStorage) Set(key string, value []byte, exp time.Duration) error {
	s._map.Set(key, &cacheEntry{
		bytes: value,
		exp:   time.Now().Add(exp),
	})
	return nil
}

func (s *memoryStorage) Get(key string) ([]byte, error) {
	val, ok := s._map.Get(key)
	if !ok {
		return nil, nil
	}

	if val.Expired() {
		s._map.Delete(key)
		s.log.Trace().Str("key", key).Msg("cache expired")
		return nil, nil
	}

	s.log.Trace().Str("key", key).Msg("returning value from cache")
	return val.bytes, nil
}

func (s *memoryStorage) Delete(key string) error {
	s._map.Delete(key)
	return nil
}

func (s *memoryStorage) Reset() error {
	s._map.Clear()
	return nil
}

func (s *memoryStorage) Close() error {
	return nil
}
