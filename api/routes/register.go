package routes

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func Register(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not present while handling register")
		return ctx.Status(500).SendString("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		slog.Error("No AuthProvider found while handling register")
		return ctx.Status(500).SendString("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
	}

	_, err := authProvider.Register(ctx)
	if err != nil {
		return ctx.Status(500).SendString("Could not register. Please try again. " + err.Error())
	}

	return ctx.Redirect("/")
}
