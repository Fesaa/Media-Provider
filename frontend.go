package main

import (
	"errors"
	"log/slog"

	middleware "github.com/Fesaa/Media-Provider/middelware"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func RegisterFrontEnd(app *fiber.App) error {

	app.Get("/", middleware.AuthHandler, home)
	app.Get("/anime", middleware.AuthHandler, anime)
	app.Get("/movies", middleware.AuthHandler, movies)

	app.Get("/login", login)

	app.Get("/status/404", status404)

	return nil
}

func status404(ctx *fiber.Ctx) error {
	return ctx.Render("404", fiber.Map{})
}

func anime(ctx *fiber.Ctx) error {
	return ctx.Render("anime", fiber.Map{})
}

func movies(ctx *fiber.Ctx) error {
	return ctx.Render("movies", fiber.Map{})
}

func home(ctx *fiber.Ctx) error {
	return ctx.Render("index", fiber.Map{})
}

func login(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("Holder not present while handling login")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		slog.Error("No AuthProvider found while handling login")
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	auth, err := authProvider.IsAuthenticated(ctx)
	if err != nil {
		slog.Error("Error while checking if user is authenticated: " + err.Error())
		return errors.New("")
	}

	if auth {
		return ctx.Redirect("/")
	}

	return ctx.Render("login", nil)
}
