package routes

import (
	"errors"
	"log/slog"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func Login(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Debug("Holder not present while handling login")
		return errors.New("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		slog.Debug("No AuthProvider found while handling login")
		return errors.New("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
	}

	err := authProvider.Login(ctx)
	if err != nil {
		return err
	}

	return ctx.Redirect(config.C.RootURL + "/")
}
