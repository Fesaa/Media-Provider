package tracing

import (
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Middleware(c *fiber.Ctx) error {
	span := trace.SpanFromContext(c.UserContext())
	requestId := contextkey.GetFromContext(c, contextkey.RequestId)

	span.SetAttributes(attribute.String("request.id", requestId))

	return c.Next()
}
