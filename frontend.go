package main

import (
	"errors"
	"log/slog"

	"github.com/Fesaa/Media-Provider/middleware"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func RegisterFrontEnd(app fiber.Router) {
	slog.Debug("Registering Front End")

	app.Get("/", middleware.AuthHandlerRedirect, home)
	app.Get("/page", middleware.AuthHandlerRedirect, page)
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
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Debug("Holder not present while handling login")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		slog.Debug("No AuthProvider found while handling login")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	auth, err := authProvider.IsAuthenticated(ctx)
	if err != nil {
		slog.Error("Error checking if user is authenticated ", "error", err)
		return errors.New("")
	}

	if auth {
		return ctx.Redirect(baseURL + "/")
	}

	return ctx.Render("login", baseURLMap)
}
