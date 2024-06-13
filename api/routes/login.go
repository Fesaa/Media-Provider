package routes

import (
	"errors"
	"github.com/Fesaa/Media-Provider/auth"
	"log/slog"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func Login(ctx *fiber.Ctx) error {
	authProvider := auth.I()
	if authProvider == nil {
		slog.Debug("No AuthProvider found while handling login")
		return errors.New("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
	}

	err := authProvider.Login(ctx)
	if err != nil {
		return err
	}

	return ctx.Redirect(config.I().GetRootURl() + "/")
}
