package frontend

import (
	middleware "github.com/Fesaa/Media-Provider/middelware"
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) error {

	app.Get("/", middleware.AuthHandler, home)
	app.Get("/search", middleware.AuthHandler, search)

	app.Get("/login", login)
	app.Get("/register", register)

	app.Get("/status/404", status404)

	return nil
}
