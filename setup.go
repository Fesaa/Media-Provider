package main

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/metadata"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/PuerkitoBio/goquery"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type appParams struct {
	dig.In

	Cfg       *config.Config
	Container *dig.Container
	Auth      auth.Provider `name:"api-key-auth"`
	Log       zerolog.Logger
}

//nolint:funlen
func ApplicationProvider(params appParams) *fiber.App {
	c := params.Container
	baseUrl := params.Cfg.BaseUrl

	app := fiber.New(fiber.Config{
		AppName: "Media-Provider",
	})

	if os.Getenv("DEV") == "" {
		app.Use(favicon.New(favicon.Config{File: "public/favicon.ico"}))
	}

	app.
		Use(requestid.New()).
		Use(recover.New(recover.Config{
			EnableStackTrace: true,
		})).
		Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:4200",
			AllowCredentials: true,
		})).
		Use(compress.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	prometheus := fiberprometheus.NewWithDefaultRegistry("media-provider")
	prometheus.RegisterAt(app, "/api/metrics", params.Auth.Middleware)
	app.Use(prometheus.Middleware)

	dontLog := []string{"/", "/api/metrics"}
	dontLogExt := []string{".js", ".html", ".css", ".svg", ".woff2", ".json"}
	httpLogger := params.Log.With().Str("handler", "http").Logger()
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &httpLogger,
		Next: func(c *fiber.Ctx) bool {
			if slices.Contains(dontLogExt, path.Ext(c.Path())) {
				return true
			}
			return slices.Contains(dontLog, c.Path()) || params.Cfg.Logging.Level > zerolog.InfoLevel
		},
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

	api.Setup(app.Group(baseUrl), c, params.Cfg, params.Log)

	app.Static(baseUrl, "./public", fiber.Static{
		Compress: true,
		MaxAge:   60 * 60,
	})

	return app
}

func RegisterCallback(app *fiber.App) {
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("./public/index.html")
	})
}

func UpdateInstalledVersion(ms services.MetadataService, log zerolog.Logger) error {
	log = log.With().Str("handler", "core").Logger()

	cur, err := ms.Get()
	if err != nil {
		return err
	}

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

	cur.Version = metadata.Version
	return ms.Update(cur)
}

func UpdateBaseUrlInIndex(cfg *config.Config, log zerolog.Logger, fs afero.Afero) error {
	baseUrl := cfg.BaseUrl
	log = log.With().Str("handler", "core").Logger()

	if os.Getenv("DEV") != "" {
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
		if os.Getenv("DOCKER") != "true" || !errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("failed to update index.html: %w", err)
		}
	} else {
		log.Info().Str("baseURL", baseUrl).Msg("Updated base URL in index.html")
	}

	return nil
}

func PprofEndPoint(log zerolog.Logger) {
	if strings.ToLower(os.Getenv("PPROF")) != "true" {
		log.Debug().Msg("pprof is not enabled, not starting http server")
		return
	}

	go func() {
		log.Info().Msg("Starting pprof endpoint on localhost:6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Error().Err(err).Msg("Failed to start pprof")
		}
	}()

}
