package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
	"strings"
	"time"
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
		ExpirationGenerator: func(ctx *fiber.Ctx, config *cache.Config) time.Duration {
			if strings.HasPrefix(ctx.Route().Path, "/api/proxy") {
				return 24 * time.Hour
			}

			return 5 * time.Minute
		},
	})

	api := app.Group("/api")

	api.Post("/login", routes.Login)

	api.Use(auth.Middleware)

	api.Post("/search", c, routes.Search)
	api.Get("/stats", routes.Stats)
	api.Post("/download", routes.Download)
	api.Post("/stop", routes.Stop)

	io := api.Group("/io")
	io.Post("/ls", routes.ListDirs)
	io.Post("/create", routes.CreateDir)

	config := api.Group("/config")
	config.Get("/", routes.GetConfig)
	config.Post("/update", routes.UpdateConfig)

	pages := config.Group("/pages")
	pages.Get("/", routes.Pages)
	pages.Get("/:index", routes.Page)
	pages.Delete("/:index", routes.RemovePage)
	pages.Post("/", routes.AddPage)
	pages.Put("/:index", routes.UpdatePage)
	pages.Post("/move", routes.MovePage)

	proxy := api.Group("/proxy", c)
	proxy.Get("/mangadex/covers/:id/:filename", routes.MangaDexCoverProxy)
}
