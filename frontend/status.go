package frontend

import "github.com/gofiber/fiber/v2"

func status404(ctx *fiber.Ctx) error {
	return ctx.Render("404", fiber.Map{})
}
