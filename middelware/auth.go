package middleware

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func AuthHandler(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("No Holder found while handling auth. Was it set before AuthHandler was registered?")
		return ctx.Status(500).SendString("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		slog.Error("No AuthProvider found while handling auth. Was it implemented in the holderImpl?")
		return ctx.Status(500).SendString("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
	}

	isAuthenticated, err := authProvider.IsAuthenticated(ctx)
	if err != nil {
		slog.Error("Error while checking authentication status %s", err)
		return ctx.Status(500).SendString("Internal Server Error. Error while checking authentication")
	}
	if !isAuthenticated {
		return ctx.Redirect("/login")
	}

	return ctx.Next()
}
