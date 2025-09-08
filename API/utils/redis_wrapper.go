package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Storage interface {
	fiber.Storage
	GetWithContext(context.Context, string) ([]byte, error)
	SetWithContext(context.Context, string, []byte, time.Duration) error
	DeleteWithContext(context.Context, string) error
	ResetWithContext(context.Context) error
}

func NewRedisCacheStorage(ctx context.Context, log zerolog.Logger, clientName, redisAddr string) Storage {
	log = log.With().Str("handler", "redis-client").Logger()

	rds := redis.NewClient(&redis.Options{
		Addr:           redisAddr,
		Password:       "",
		DB:             0,
		ClientName:     clientName,
		IdentitySuffix: "media-provider",
	})

	if err := rds.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}

	return &redisWrapper{
		rdb: rds,
		log: log,
	}
}

type redisWrapper struct {
	rdb *redis.Client
	log zerolog.Logger
}

func (r *redisWrapper) Get(key string) ([]byte, error) {
	return r.rdb.Get(context.Background(), key).Bytes()
}

func (r *redisWrapper) GetWithContext(ctx context.Context, key string) ([]byte, error) {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesCache,
		trace.WithAttributes(
			attribute.String("cache.operation", "get"),
			attribute.String("cache.key", key),
		))
	defer span.End()

	b, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		r.log.Trace().Err(err).Str("key", key).Msg("failed to get")
	}
	return b, err
}

func (r *redisWrapper) Set(key string, value []byte, expiration time.Duration) error {
	return r.rdb.Set(context.Background(), key, value, expiration).Err()
}

func (r *redisWrapper) SetWithContext(ctx context.Context, key string, val []byte, exp time.Duration) error {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesCache,
		trace.WithAttributes(
			attribute.String("cache.operation", "set"),
			attribute.String("cache.key", key),
			attribute.Int("cache.length", len(val)),
			attribute.String("cache.expiration", fmt.Sprintf("%v", exp)),
		))
	defer span.End()

	return r.logAndReturn(r.rdb.Set(ctx, key, val, exp).Err(), key, "failed to set")
}

func (r *redisWrapper) Delete(key string) error {
	return r.rdb.Del(context.Background(), key).Err()
}

func (r *redisWrapper) DeleteWithContext(ctx context.Context, key string) error {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesCache,
		trace.WithAttributes(
			attribute.String("cache.operation", "delete"),
			attribute.String("cache.key", key),
		))
	defer span.End()
	return r.logAndReturn(r.rdb.Del(ctx, key).Err(), key, "failed to delete")
}

func (r *redisWrapper) Reset() error {
	return r.ResetWithContext(context.Background())
}

func (r *redisWrapper) ResetWithContext(ctx context.Context) error {
	return r.logAndReturn(r.rdb.FlushAll(ctx).Err(), "", "failed to reset")
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
