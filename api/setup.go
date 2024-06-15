package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

func Setup(app fiber.Router) {
	slog.Debug("Registering api routes")
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	api.Post("/login", routes.Login)
	api.Get("/logout", routes.Logout)

	api.Post("/search", auth.Middleware(), routes.Search)
	api.Get("/stats", auth.Middleware(), routes.Stats)
	api.Post("/download/", auth.Middleware(), routes.Download)
	api.Post("/stop/", auth.Middleware(), routes.Stop)

	api.Get("/pages", auth.Middleware(), routes.Pages)
	api.Get("/pages/:index", auth.Middleware(), routes.Page)

	io := api.Group("/io")
	io.Post("/ls", auth.Middleware(), routes.ListDirs)
	io.Post("/create", auth.Middleware(), routes.CreateDir)

	config := api.Group("/config")
	config.Get("/", auth.Middleware(), routes.GetConfig)
}
