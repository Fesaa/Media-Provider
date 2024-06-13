package auth

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

func Middleware(redirect ...bool) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		isAuthenticated, err := authProvider.IsAuthenticated(ctx)
		if err != nil {
			slog.Error("Error while checking authentication status %s", err)
			return ctx.Status(500).SendString("Internal Server Error. Error while checking authentication")
		}
		if !isAuthenticated {
			if len(redirect) > 0 && redirect[0] {
				return ctx.Redirect(config.I().GetRootURl() + "/login")
			}
			return ctx.Status(401).SendString("Unauthorized")
		}

		return ctx.Next()
	}
}
