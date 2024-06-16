package main

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"

	"github.com/gofiber/fiber/v2"
)

func RegisterFrontEnd(app fiber.Router) {
	log.Debug("Registering Front End")

	app.Get("/", auth.Middleware(true), home)
	app.Get("/page", auth.Middleware(true), page)
	app.Get("/login", login)

	app.Get("/status/404", status404)
}

func status404(ctx *fiber.Ctx) error {
	return ctx.Render("404", baseURLMap)
}

func page(ctx *fiber.Ctx) error {
	return ctx.Render("page", baseURLMap)
}

func home(ctx *fiber.Ctx) error {
	return ctx.Render("index", baseURLMap)
}

func login(ctx *fiber.Ctx) error {
	success, err := auth.I().IsAuthenticated(ctx)
	if err != nil {
		log.Error("Error checking if user is authenticated ", "error", err)
		return err
	}

	if success {
		return ctx.Redirect(baseURL + "/")
	}

	return ctx.Render("login", baseURLMap)
}
