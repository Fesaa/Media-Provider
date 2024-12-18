package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
	"strings"
	"time"
)

func Setup(app fiber.Router, db *db.Database) {
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
	routes.RegisterUserRoutes(api, db, c)
	routes.RegisterProxyRoutes(api, db, c)
	routes.RegisterContentRoutes(api, db, c)
	routes.RegisterIoRoutes(api, db, c)
	routes.RegisterConfigRoutes(api, db, c)
	routes.RegisterPageRoutes(api, db, c)
	routes.RegisterSubscriptionRoutes(api, db, c)
}
