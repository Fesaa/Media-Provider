package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func Login(ctx *fiber.Ctx) error {
	authProvider := auth.I()
	err := authProvider.Login(ctx)
	if err != nil {
		return err
	}

	return ctx.Redirect(config.Get(ctx).BaseUrl + "/")
}
