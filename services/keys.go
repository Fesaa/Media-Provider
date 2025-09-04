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

func GetFromContext[T any](ctx *fiber.Ctx, key ContextKey[T]) T {
	return ctx.Locals(string(key)).(T)
}
