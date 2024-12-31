package api

import (
	"context"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"time"
)

func cacheStorage(cfg *config.Config, log zerolog.Logger) fiber.Storage {
	switch cfg.Cache.Type {
	case config.REDIS:
		return newRedisCacheStorage(cfg, log)
	case config.MEMORY:
		return nil
	default:
		// the fiber cache config falls back to memory on its own
		return nil
	}
}

func newRedisCacheStorage(cfg *config.Config, log zerolog.Logger) fiber.Storage {
	rds := redis.NewClient(&redis.Options{
		Addr:           cfg.Cache.RedisAddr,
		Password:       "",
		DB:             0,
		ClientName:     "go-fiber-storage",
		IdentitySuffix: "media-provider",
	})

	if err := rds.Ping(context.Background()).Err(); err != nil {
		log.Warn().Err(err).Msg("failed to connect to redis")
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
