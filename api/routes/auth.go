package routes

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/gofiber/fiber/v2"
)

func Login(ctx *fiber.Ctx) error {
	authProvider := auth.I()
	res, err := authProvider.Login(ctx)
	if err != nil {
		return err
	}

	return ctx.JSON(res)
}
