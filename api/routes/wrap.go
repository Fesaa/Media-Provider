package routes

import (
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

func wrap(f func(l *log.Logger, ctx *fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		requestID := ctx.Locals("requestid").(string)
		l := log.With(slog.String("request-id", requestID))
		return f(l, ctx)
	}
}
