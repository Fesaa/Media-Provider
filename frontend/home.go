package frontend

import "github.com/gofiber/fiber/v2"

func home(ctx *fiber.Ctx) error {
	return ctx.Render("index", fiber.Map{})
}
