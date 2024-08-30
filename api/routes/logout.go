package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func Logout(ctx *fiber.Ctx) error {
	authProvider := auth.I()
	err := authProvider.Logout(ctx)
	if err != nil {
		return ctx.Status(500).SendString("Could not logout. Please try again. " + err.Error())
	}

	return ctx.Redirect(config.Get(ctx).BaseUrl + "/login")
}
