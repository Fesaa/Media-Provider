package middleware

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func AuthHandler(ctx *fiber.Ctx) error {
	return authHandlerFactory(false)(ctx)
}

func AuthHandlerRedirect(ctx *fiber.Ctx) error {
	return authHandlerFactory(true)(ctx)
}

func authHandlerFactory(redirect bool) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
		if !ok {
			slog.Debug("No Holder found while handling auth. Was it set before AuthHandler was registered?")
			return ctx.Status(500).SendString("Internal Server Error.\nHolder was not present. Please contact the administrator.")
		}

		authProvider := holder.GetAuthProvider()
		if authProvider == nil {
			slog.Debug("No AuthProvider found while handling auth. Was it implemented in the holderImpl?")
			return ctx.Status(500).SendString("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
		}

		isAuthenticated, err := authProvider.IsAuthenticated(ctx)
		if err != nil {
			slog.Error("Error while checking authentication status %s", err)
			return ctx.Status(500).SendString("Internal Server Error. Error while checking authentication")
		}
		if !isAuthenticated {
			if redirect {
				return ctx.Redirect(config.C.RootURL + "/login")
			}
			return ctx.Status(401).SendString("Unauthorized")
		}

		return ctx.Next()
	}
}
