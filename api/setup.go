package api

import (
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	utils2 "github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

func Setup(router fiber.Router, container *dig.Container, settingsService services.SettingsService, log zerolog.Logger) error {
	log.Debug().Str("handler", "http-routing").Msg("registering api routes")

	settings, err := settingsService.GetSettingsDto()
	if err != nil {
		return err
	}

	cacheHandler := cache.New(cache.Config{
		Storage:      cacheStorage(settings, log),
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
				if settings.CacheType == config.REDIS {
					return 7 * 24 * time.Hour
				}
				return 24 * time.Hour
			}

			return c.Expiration
		},
	})

	scope := container.Scope("mp::http::api")

	utils2.Must(scope.Decorate(utils2.Identity(log.With().Str("handler", "api").Logger())))
	utils2.Must(scope.Provide(utils2.Identity(router.Group("/api"))))
	utils2.Must(scope.Provide(utils2.Identity(cacheHandler), dig.Name("cache")))

	utils2.Must(scope.Invoke(routes.RegisterUserRoutes))
	utils2.Must(scope.Invoke(routes.RegisterProxyRoutes))
	utils2.Must(scope.Invoke(routes.RegisterContentRoutes))
	utils2.Must(scope.Invoke(routes.RegisterIoRoutes))
	utils2.Must(scope.Invoke(routes.RegisterConfigRoutes))
	utils2.Must(scope.Invoke(routes.RegisterPageRoutes))
	utils2.Must(scope.Invoke(routes.RegisterSubscriptionRoutes))
	utils2.Must(scope.Invoke(routes.RegisterPreferencesRoutes))
	utils2.Must(scope.Invoke(routes.RegisterNotificationRoutes))
	utils2.Must(scope.Invoke(routes.RegisterMetadataRoutes))

	return nil
}

func cacheStorage(settings payload.Settings, log zerolog.Logger) fiber.Storage {
	switch settings.CacheType {
	case config.REDIS:
		return utils2.NewRedisCacheStorage(log, "go-fiber-http-cache", settings.RedisAddr)
	case config.MEMORY:
		return nil
	default:
		// the fiber cache config falls back to memory on its own
		return nil
	}
}
