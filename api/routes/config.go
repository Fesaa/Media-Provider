package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func GetConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(config.I())
}
