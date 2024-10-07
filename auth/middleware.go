package auth

import (
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
)

func Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := authProvider.IsAuthenticated(ctx)
	if err != nil {
		log.Error("Error while checking authentication status", "err", err)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	if !isAuthenticated {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	return ctx.Next()
}
