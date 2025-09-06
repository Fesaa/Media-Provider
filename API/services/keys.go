package services

import (
	"fmt"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

// ContextKey is a value type aware abstraction around fiber.Ctx.locals
//
// Golang: PLEASE allow generics on struct methods
type ContextKey[T any] string

var (
	UserKey            = ContextKey[models.User]("user")
	ServiceProviderKey = ContextKey[*dig.Container]("service-provider")
	LoggerKey          = ContextKey[zerolog.Logger]("logger")
	RequestIdKey       = ContextKey[string]("requestid")
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

// GetFromContext returns the value of the context, panics if no present
func GetFromContext[T any](ctx *fiber.Ctx, key ContextKey[T]) T {
	value, ok := GetFromContextSafe(ctx, key)
	if !ok {
		panic(fmt.Errorf("key %s not found in context", key.Value()))
	}

	return value
}

// GetFromContextSafe returns the value of the context, or zero, false if not present
func GetFromContextSafe[T any](ctx *fiber.Ctx, key ContextKey[T]) (T, bool) {
	value, ok := ctx.Locals(key.Value()).(T)
	return value, ok
}
