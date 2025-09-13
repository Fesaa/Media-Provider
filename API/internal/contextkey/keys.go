package contextkey

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

// ContextKey is a value type aware abstraction around fiber.Ctx.locals
//
// Golang: PLEASE allow generics on struct methods
type ContextKey[T any] string

var (
	User            = ContextKey[models.User]("user")
	ServiceProvider = ContextKey[*dig.Container]("service-provider")
	Logger          = ContextKey[zerolog.Logger]("logger")
	RequestId       = ContextKey[string]("requestid")
)
