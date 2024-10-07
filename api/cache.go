package api

import (
	"context"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"time"
)

func cacheStorage() fiber.Storage {
	switch config.I().Cache.Type {
	case config.REDIS:
		return newRedisCacheStorage()
	default:
		// the fiber cache config falls back to memory on its own
		return nil
	}
}

func newRedisCacheStorage() fiber.Storage {
	rds := redis.NewClient(&redis.Options{
		Addr:           config.I().Cache.RedisAddr,
		Password:       "",
		DB:             0,
		ClientName:     "go-fiber-storage",
		IdentitySuffix: "media-provider",
	})

	if err := rds.Ping(context.Background()).Err(); err != nil {
		log.Warn("Cannot connect to redis server, falling back", "err", err)
		return nil
	}

	return &redisWrapper{rdb: rds}
}

type redisWrapper struct {
	rdb *redis.Client
}

func (r *redisWrapper) Get(key string) ([]byte, error) {
	return r.rdb.Get(context.Background(), key).Bytes()
}

func (r *redisWrapper) Set(key string, val []byte, exp time.Duration) error {
	return r.rdb.Set(context.Background(), key, val, exp).Err()
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
