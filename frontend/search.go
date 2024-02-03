package frontend

import "github.com/gofiber/fiber/v2"

func search(ctx *fiber.Ctx) error {
	return ctx.Render("search", fiber.Map{})
}
