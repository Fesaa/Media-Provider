package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
	"log/slog"
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

	api.Post("/login", logWrap(routes.LoginUser))
	api.Post("/register", logWrap(routes.RegisterUser))
	api.Get("/any-user-exists", routes.AnyUserExists)

	proxy := api.Group("/proxy", c)
	proxy.Get("/mangadex/covers/:id/:filename", auth.MiddlewareWithApiKey, routes.MangaDexCoverProxy)
	proxy.Get("/webtoon/covers/:date/:id/:filename", auth.MiddlewareWithApiKey, routes.WebToonCoverProxy)

	api.Use(auth.Middleware)

	user := api.Group("/user")
	user.Get("/refresh-api-key", logWrap(routes.RefreshApiKey))

	api.Post("/search", c, routes.Search)
	api.Get("/stats", routes.Stats)
	api.Post("/download", routes.Download)
	api.Post("/stop", routes.Stop)

	io := api.Group("/io")
	io.Post("/ls", routes.ListDirs)
	io.Post("/create", routes.CreateDir)

	configGroup := api.Group("/config")
	configGroup.Get("/", routes.GetConfig)
	configGroup.Post("/update", routes.UpdateConfig)

	pages := api.Group("/pages")
	pages.Get("/", logWrap(routes.Pages))
	pages.Get("/:index", logWrap(routes.Page))
	pages.Post("/upsert", logWrap(routes.UpsertPage))
	pages.Delete("/:pageId", logWrap(routes.DeletePage))
	pages.Post("/swap", logWrap(routes.SwapPage))
}

func logWrap(f func(l *log.Logger, ctx *fiber.Ctx) error) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		requestID := ctx.Locals("requestid").(string)
		l := log.With(slog.String("request-id", requestID))
		return f(l, ctx)
	}
}
