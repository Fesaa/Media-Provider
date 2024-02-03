package routes

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func Logout(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not present while handling logout")
		return ctx.Status(500).SendString("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		slog.Error("No AuthProvider found while handling logout")
		return ctx.Status(500).SendString("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
	}

	err := authProvider.Logout(ctx)
	if err != nil {
		return ctx.Status(500).SendString("Could not logout. Please try again. " + err.Error())
	}

	return ctx.Redirect("/login")
}
