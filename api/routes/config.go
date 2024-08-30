package routes

import (
	"github.com/gofiber/fiber/v2"
)

func GetConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(ctx.Locals("cfg"))
}
