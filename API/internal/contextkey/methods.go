package contextkey

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Value returns the string value of the ContextKey, this should be used when setting or getting
// from fiber.Ctx locals
func (ctx ContextKey[T]) Value() string {
	return string(ctx)
}

// SetInContext sets the value in fiber.Ctx.Locals such that GetFromContext returns it when passing the same key
func SetInContext[T any](ctx *fiber.Ctx, key ContextKey[T], value T) {
	ctx.Locals(key.Value(), value)
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
