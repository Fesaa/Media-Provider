package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
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
		Storage:      cacheStorage(),
		CacheControl: true,
		Next: func(c *fiber.Ctx) bool {
			return false
		},
		KeyGenerator: func(ctx *fiber.Ctx) string {
			if ctx.Method() == fiber.MethodPost {
				return utils.CopyString(ctx.Path()) + "_" + string(utils.CopyBytes(ctx.Body()))
			}
			return utils.CopyString(ctx.Path())
		},
		Methods:    []string{fiber.MethodGet, fiber.MethodPost},
		Expiration: time.Hour,
		ExpirationGenerator: func(ctx *fiber.Ctx, c *cache.Config) time.Duration {
			if strings.HasPrefix(ctx.Route().Path, "/api/proxy") {
				if config.I().Cache.Type == config.REDIS {
					return 7 * 24 * time.Hour
				}
				return 24 * time.Hour
			}

			return c.Expiration
		},
	})

	api := app.Group("/api")

	api.Post("/login", routes.Login)

	proxy := api.Group("/proxy", c)
	proxy.Get("/mangadex/covers/:id/:filename", auth.MiddlewareWithApiKey, routes.MangaDexCoverProxy)
	proxy.Get("/webtoon/covers/:date/:id/:filename", auth.MiddlewareWithApiKey, routes.WebToonCoverProxy)

	api.Use(auth.Middleware)

	api.Post("/search", c, routes.Search)
	api.Get("/stats", routes.Stats)
	api.Post("/download", routes.Download)
	api.Post("/stop", routes.Stop)

	io := api.Group("/io")
	io.Post("/ls", routes.ListDirs)
	io.Post("/create", routes.CreateDir)

	configGroup := api.Group("/config")
	configGroup.Get("/", routes.GetConfig)
	configGroup.Get("/refresh-api-key", routes.RefreshApiKey)
	configGroup.Post("/update", routes.UpdateConfig)

	pages := configGroup.Group("/pages")
	pages.Get("/", routes.Pages)
	pages.Get("/:index", routes.Page)
	pages.Delete("/:index", routes.RemovePage)
	pages.Post("/", routes.AddPage)
	pages.Put("/:index", routes.UpdatePage)
	pages.Post("/move", routes.MovePage)
}
