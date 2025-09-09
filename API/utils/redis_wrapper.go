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
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanSetupService,
		trace.WithAttributes(tracing.WithServiceName("RedisCacheStorage")))
	defer span.End()
	log = log.With().Str("handler", "redis-client").Logger()

	rds := redis.NewClient(&redis.Options{
		Addr:           redisAddr,
		Password:       "",
		DB:             0,
		ClientName:     clientName,
		IdentitySuffix: "media-provider",
	})

	if err := rds.Ping(ctx).Err(); err != nil {
		span.RecordError(err)
		span.End()                                             // fatal exists, manually end the span
		log.Fatal().Err(err).Msg("failed to connect to redis") //nolint: gocritic
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
	if err != nil {
		span.SetAttributes(attribute.String("cache.error", err.Error()))
		if errors.Is(err, redis.Nil) {
			span.SetAttributes(attribute.Bool("cache.miss", true))
		} else {
			r.log.Trace().Err(err).Str("key", key).Msg("failed to get")
		}
	} else {
		span.SetAttributes(attribute.Int("cache.result_size", len(b)))
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

	err := r.rdb.Set(ctx, key, val, exp).Err()
	if err != nil {
		span.SetAttributes(attribute.String("cache.error", err.Error()))
		r.log.Trace().Err(err).Str("key", key).Msg("failed to set")
	} else {
		span.SetAttributes(attribute.Bool("cache.success", true))
	}
	return err
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

	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		span.SetAttributes(attribute.String("cache.error", err.Error()))
		r.log.Trace().Err(err).Str("key", key).Msg("failed to delete")
	} else {
		span.SetAttributes(attribute.Bool("cache.success", true))
	}
	return err
}

func (r *redisWrapper) Reset() error {
	return r.ResetWithContext(context.Background())
}

func (r *redisWrapper) ResetWithContext(ctx context.Context) error {
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanServicesCache,
		trace.WithAttributes(
			attribute.String("cache.operation", "reset"),
		))
	defer span.End()

	err := r.rdb.FlushAll(ctx).Err()
	if err != nil {
		span.SetAttributes(attribute.String("cache.error", err.Error()))
		r.log.Trace().Err(err).Msg("failed to reset")
	} else {
		span.SetAttributes(attribute.Bool("cache.success", true))
	}
	return err
}

func (r *redisWrapper) Close() error {
	return r.rdb.Close()
}
