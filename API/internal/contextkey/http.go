package contextkey

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

func Middleware(container *dig.Container, log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		SetInContext(c, ServiceProvider, container)

		requestId := GetFromContext(c, RequestId)
		log = log.With().Str(RequestId.Value(), requestId).Logger()
		SetInContext(c, Logger, log)

		return c.Next()
	}
}
