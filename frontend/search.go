package frontend

import "github.com/gofiber/fiber/v2"

func anime(ctx *fiber.Ctx) error {
	return ctx.Render("anime", fiber.Map{})
}

func movies(ctx *fiber.Ctx) error {
	return ctx.Render("movies", fiber.Map{})
}
