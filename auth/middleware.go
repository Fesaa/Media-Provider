package auth

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
)

func Middleware(redirect ...bool) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		isAuthenticated, err := authProvider.IsAuthenticated(ctx)
		if err != nil {
			log.Error("Error while checking authentication status %s", err)
			return ctx.Status(500).SendString("Internal Server Error. Error while checking authentication")
		}
		if !isAuthenticated {
			if len(redirect) > 0 && redirect[0] {
				return ctx.Redirect(config.I().BaseUrl + "/login")
			}
			return ctx.Status(401).SendString("Unauthorized")
		}

		return ctx.Next()
	}
}
