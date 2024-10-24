package auth

import (
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
)

func Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := jwtProvider.IsAuthenticated(ctx)
	if err != nil {
		log.Error("Error while checking authentication status", "err", err)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	if !isAuthenticated {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	return ctx.Next()
}

// MiddlewareWithApiKey Allows apiKeys to be used to authenticate, will always fall back to JWT tokens, if not.
func MiddlewareWithApiKey(ctx *fiber.Ctx) error {
	isAuthenticated, err := apiKeyProvider.IsAuthenticated(ctx)
	if err != nil {
		log.Warn("error while checking api key auth", "err", err)
	}
	if !isAuthenticated {
		return Middleware(ctx)
	}
	return ctx.Next()
}
