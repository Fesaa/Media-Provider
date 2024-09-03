package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
)

func Setup(app fiber.Router) {
	log.Debug("registering api routes")

	c := cache.New(cache.Config{
		CacheControl: true,
		Next: func(c *fiber.Ctx) bool {
			return false
		},
		KeyGenerator: func(ctx *fiber.Ctx) string {
			return ctx.Path() + "_" + string(utils.CopyBytes(ctx.Body()))
		},
		Methods: []string{fiber.MethodGet, fiber.MethodPost},
	})

	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	api.Post("/login", routes.Login)
	api.Get("/logout", routes.Logout)
	api.Post("/update-password", auth.Middleware(), routes.UpdatePassword)

	api.Post("/search", auth.Middleware(), c, routes.Search)
	api.Get("/stats", auth.Middleware(), routes.Stats)
	api.Post("/download", auth.Middleware(), routes.Download)
	api.Post("/stop", auth.Middleware(), routes.Stop)

	io := api.Group("/io")
	io.Post("/ls", auth.Middleware(), routes.ListDirs)
	io.Post("/create", auth.Middleware(), routes.CreateDir)

	config := api.Group("/config")
	config.Get("/", auth.Middleware(), routes.GetConfig)

	pages := config.Group("/pages")
	pages.Get("/", auth.Middleware(), routes.Pages)
	pages.Get("/:index", auth.Middleware(), routes.Page)
	pages.Delete("/:index", auth.Middleware(), routes.RemovePage)
	pages.Post("/", auth.Middleware(), routes.AddPage)
	pages.Put("/:index", auth.Middleware(), routes.UpdatePage)
	pages.Post("/move", auth.Middleware(), routes.MovePage)
}
