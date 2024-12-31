package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/config"
	utils2 "github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"strings"
	"time"
)

func Setup(router fiber.Router, container *dig.Container, cfg *config.Config, log zerolog.Logger) {
	log.Debug().Msg("registering api routes")

	cacheHandler := cache.New(cache.Config{
		Storage:      cacheStorage(cfg, log),
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
				if cfg.Cache.Type == config.REDIS {
					return 7 * 24 * time.Hour
				}
				return 24 * time.Hour
			}

			return c.Expiration
		},
	})

	scope := container.Scope("mp::http::api")

	utils2.Must(scope.Decorate(utils2.Identity(log.With().Str("handler", "http").Logger())))
	utils2.Must(scope.Provide(utils2.Identity(router.Group("/api"))))
	utils2.Must(scope.Provide(utils2.Identity(cacheHandler), dig.Name("cache")))

	utils2.Must(scope.Invoke(routes.RegisterUserRoutes))
	utils2.Must(scope.Invoke(routes.RegisterProxyRoutes))
	utils2.Must(scope.Invoke(routes.RegisterContentRoutes))
	utils2.Must(scope.Invoke(routes.RegisterIoRoutes))
	utils2.Must(scope.Invoke(routes.RegisterConfigRoutes))
	utils2.Must(scope.Invoke(routes.RegisterPageRoutes))
	utils2.Must(scope.Invoke(routes.RegisterSubscriptionRoutes))
}
