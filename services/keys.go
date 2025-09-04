package services

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type ContextKey[T any] string

var (
	UserKey            = ContextKey[models.User]("user")
	ServiceProviderKey = ContextKey[*dig.Container]("service-provider")
)

// Value returns the string value of the ContextKey, this should be used when setting or getting
// from fiber.Ctx locals
func (ctx ContextKey[T]) Value() string {
	return string(ctx)
}

func GetFromContext[T any](ctx *fiber.Ctx, key ContextKey[T]) T {
	return ctx.Locals(key.Value()).(T)
}
