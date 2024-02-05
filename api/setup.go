package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	middleware "github.com/Fesaa/Media-Provider/middelware"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, holder models.Holder) error {
	api := app.Group("/api")

	api.Post("/login", routes.Login)
	api.Get("/logout", routes.Logout)

	api.Post("/search", middleware.AuthHandler, routes.Search)
	api.Get("/stats", middleware.AuthHandler, routes.Stats)
	api.Get("/download/:infoHash", middleware.AuthHandler, routes.Download)
	api.Get("/stop/:infoHash", middleware.AuthHandler, routes.Stop)

	return nil
}
