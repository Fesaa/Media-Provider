package contextkey

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Value returns the string value of the ContextKey, this should be used when setting or getting
// from fiber.Ctx locals
func (ctx ContextKey[T]) Value() string {
	return string(ctx)
}

// SetInContext sets the value in fiber.Ctx.Locals such that GetFromContext returns it when passing the same key
// The fiber.Ctx.UserContext is also updated with the same value
func SetInContext[T any](ctx *fiber.Ctx, key ContextKey[T], value T) {
	ctx.Locals(key.Value(), value)
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), key.Value(), value))
}

// GetFromContextSafe returns the value of the context, or zero, false if not present
func GetFromContextSafe[T any](ctx *fiber.Ctx, key ContextKey[T]) (T, bool) {
	value, ok := ctx.Locals(key.Value()).(T)
	return value, ok
}

// GetFromContext returns the value of the context, panics if no present
func GetFromContext[T any](ctx *fiber.Ctx, key ContextKey[T]) T {
	value, ok := GetFromContextSafe(ctx, key)
	if !ok {
		panic(fmt.Errorf("key %s not found in context", key.Value()))
	}

	return value
}

func GetFromCtxSafe[T any](ctx context.Context, key ContextKey[T]) (T, bool) {
	value, ok := ctx.Value(key.Value()).(T)
	return value, ok
}

func GetFromCtx[T any](ctx context.Context, key ContextKey[T]) T {
	value, ok := GetFromCtxSafe(ctx, key)
	if !ok {
		panic(fmt.Errorf("key %s not found in context", key.Value()))
	}
	return value
}

func GetFromCtxOrDefault[T any](ctx context.Context, key ContextKey[T], def ...T) T {
	value, ok := GetFromCtxSafe(ctx, key)
	if !ok {
		if len(def) > 0 {
			return def[0]
		}
		var zero T
		return zero
	}
	return value
}
