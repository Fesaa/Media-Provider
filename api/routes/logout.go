package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"log/slog"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func Logout(ctx *fiber.Ctx) error {
	authProvider := auth.I()
	if authProvider == nil {
		slog.Debug("No AuthProvider found while handling logout")
		return ctx.Status(500).SendString("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
	}

	err := authProvider.Logout(ctx)
	if err != nil {
		return ctx.Status(500).SendString("Could not logout. Please try again. " + err.Error())
	}

	return ctx.Redirect(config.I().GetRootURl() + "/login")
}
