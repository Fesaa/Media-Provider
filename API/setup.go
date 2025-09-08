package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/metadata"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	fiberutils "github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"
)

type appParams struct {
	dig.In

	Cfg       *config.Config
	Container *dig.Container
	Auth      services.AuthMiddleware
	Log       zerolog.Logger
}

//nolint:funlen
func applicationProvider(params appParams) *fiber.App {
	c := params.Container
	baseUrl := params.Cfg.BaseUrl

	app := fiber.New(fiber.Config{
		AppName:                 "Media-Provider",
		DisableStartupMessage:   true,
		ErrorHandler:            routes.ErrorHandler,
		EnableTrustedProxyCheck: len(config.TrustedIps) > 0,
		TrustedProxies:          config.TrustedIps,
	})

	if !config.Development {
		app.Use(favicon.New(favicon.Config{File: "public/favicon.ico"}))
	}

	app.
		Use(otelfiber.Middleware(otelfiber.WithServerName(metadata.Identifier))).
		Use(limiter.New(limiter.Config{
			Max:               1000,
			Expiration:        time.Minute,
			LimiterMiddleware: limiter.SlidingWindow{},
		})).
		Use(requestid.New(requestid.Config{
			ContextKey: contextkey.RequestId.Value(),
			Generator:  fiberutils.UUIDv4,
		})).
		Use(encryptcookie.New(encryptcookie.Config{
			Key: params.Cfg.CookieSecret,
		})).
		Use(recover.New(recover.Config{
			EnableStackTrace: true,
		})).
		Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:4400",
			AllowCredentials: true,
		})).
		Use(compress.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	prometheus := fiberprometheus.NewWithDefaultRegistry("media-provider")
	prometheus.RegisterAt(app, "/api/metrics", params.Auth.Middleware)
	app.Use(prometheus.Middleware)

	httpLogger := params.Log.With().Str("handler", "http").Logger()
	if !config.NoHttpLog {
		dontLog := []string{"/", "/api/metrics"}
		dontLogExt := []string{".js", ".html", ".css", ".svg", ".woff2", ".json"}

		app.Use(fiberzerolog.New(fiberzerolog.Config{
			Logger: &httpLogger,
			Next: func(c *fiber.Ctx) bool {
				if slices.Contains(dontLogExt, path.Ext(c.Path())) {
					return true
				}
				return slices.Contains(dontLog, c.Path()) || params.Cfg.Logging.Level > zerolog.InfoLevel
			},
			Levels: func() []zerolog.Level {
				if config.ReducedHttpLog {
					return []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.TraceLevel}
				}
				return []zerolog.Level{zerolog.ErrorLevel, zerolog.WarnLevel, zerolog.InfoLevel}
			}(),
			Fields: []string{
				fiberzerolog.FieldUserAgent,
				fiberzerolog.FieldIP,
				fiberzerolog.FieldLatency,
				fiberzerolog.FieldStatus,
				fiberzerolog.FieldMethod,
				fiberzerolog.FieldURL,
				fiberzerolog.FieldError,
				fiberzerolog.FieldRequestID,
			},
		}))
	}

	app.Use(contextkey.Middleware(params.Container, httpLogger)).
		Use(MiddlewareTracingSetRequestId)

	scope := c.Scope("init::api")
	utils.Must(scope.Provide(utils.Identity(app.Group(baseUrl))))
	utils.Must(scope.Invoke(api.Setup))

	app.Static(baseUrl, "./public", fiber.Static{
		Compress: true,
		MaxAge:   60 * 60,
	})

	return app
}

func registerCallback(app *fiber.App) {
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})
}

func updateInstalledVersion(ss services.SettingsService, log zerolog.Logger, ctx context.Context) error {
	ctx, span := tracing.TracerMain.Start(ctx, tracing.SpanUpdateVersion)
	defer span.End()

	log = log.With().Str("handler", "core").Logger()

	settings, err := ss.GetSettingsDto(ctx)
	if err != nil {
		return err
	}

	cur := settings.Metadata

	if cur.Version.Equal(metadata.Version) {
		log.Trace().Msg("no version changes")
		return nil
	}

	if cur.Version.Newer(metadata.Version) {
		log.Warn().
			Str("installedVersion", cur.Version.String()).
			Str("actualVersion", metadata.Version.String()).
			Msg("Installed version is newer, want is going on? Bringing back to sync!")
	}
	return ss.UpdateCurrentVersion(ctx)
}

func updateBaseUrlInIndex(cfg *config.Config, log zerolog.Logger, fs afero.Afero) error {
	baseUrl := cfg.BaseUrl
	log = log.With().Str("handler", "core").Logger()

	if config.Development {
		log.Debug().Msg("Skipping base url update in DEV environment")
		return nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	indexHtmlPath := filepath.Join(cwd, "public", "index.html")

	file, err := fs.Open(indexHtmlPath)
	if err != nil {
		return err
	}
	defer func(file afero.File) {
		if err = file.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close index.html")
		}
	}(file)

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		return fmt.Errorf("failed to load index.html: %w", err)
	}

	doc.Find("head base").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("href", baseUrl)
	})

	html, err := doc.Html()
	if err != nil {
		return fmt.Errorf("error converting document to HTML: %w", err)
	}

	err = fs.WriteFile(indexHtmlPath, []byte(html), 0644)
	if err != nil {
		// Ignore errors when running as non-root in docker
		if !config.Docker || !errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("failed to update index.html: %w", err)
		}
	} else {
		log.Info().Str("baseURL", baseUrl).Msg("Updated base URL in index.html")
	}

	return nil
}

func MiddlewareTracingSetRequestId(c *fiber.Ctx) error {
	span := trace.SpanFromContext(c.UserContext())
	requestId := contextkey.GetFromContext(c, contextkey.RequestId)

	span.SetAttributes(attribute.String("request.id", requestId))

	return c.Next()
}
