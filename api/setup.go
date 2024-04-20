package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/middleware"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func Setup(app fiber.Router, holder models.Holder) {
	api := app.Group("/api")

	api.Post("/login", routes.Login)
	api.Get("/logout", routes.Logout)

	api.Post("/search", middleware.AuthHandler, routes.Search)
	api.Get("/stats", middleware.AuthHandler, routes.Stats)
	api.Post("/download/", middleware.AuthHandler, routes.Download)
	api.Get("/stop/:infoHash", middleware.AuthHandler, routes.Stop)

	api.Get("/pages", middleware.AuthHandler, routes.Pages)
	api.Get("/pages/:index", middleware.AuthHandler, routes.Page)

	io := api.Group("/io")
	io.Post("/ls", middleware.AuthHandler, routes.ListDirs)
	io.Post("/create", middleware.AuthHandler, routes.CreateDir)

	config := api.Group("/config")
	config.Post("/reload", middleware.AuthHandler, routes.ReloadPages)
	config.Get("/", middleware.AuthHandler, routes.GetConfig)
}
