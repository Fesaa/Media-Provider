package utils

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"time"
)

func NewRedisCacheStorage(log zerolog.Logger, clientName, redisAddr string) fiber.Storage {
	rds := redis.NewClient(&redis.Options{
		Addr:           redisAddr,
		Password:       "",
		DB:             0,
		ClientName:     clientName,
		IdentitySuffix: "media-provider",
	})

	if err := rds.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	return &redisWrapper{
		rdb: rds,
		log: log.With().Str("handler", "redis-client").Logger(),
	}
}

type redisWrapper struct {
	rdb *redis.Client
	log zerolog.Logger
}

func (r *redisWrapper) Get(key string) ([]byte, error) {
	b, err := r.rdb.Get(context.Background(), key).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		r.log.Trace().Err(err).Str("key", key).Msg("failed to get")
	}
	return b, err
}

func (r *redisWrapper) Set(key string, val []byte, exp time.Duration) error {
	return r.logAndReturn(r.rdb.Set(context.Background(), key, val, exp).Err(), key, "failed to set")
}

func (r *redisWrapper) Delete(key string) error {
	return r.logAndReturn(r.rdb.Del(context.Background(), key).Err(), key, "failed to delete")
}

func (r *redisWrapper) Reset() error {
	return r.logAndReturn(r.rdb.FlushAll(context.Background()).Err(), "", "failed to reset")
}

func (r *redisWrapper) Close() error {
	return r.rdb.Close()
}

func (r *redisWrapper) logAndReturn(err error, key, msg string) error {
	if err != nil && !errors.Is(err, redis.Nil) {
		r.log.Trace().Err(err).Str("key", key).Msg(msg)
	}
	return err
}
