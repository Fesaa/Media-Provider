package main

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func SetupApp(baseUrl string) *fiber.App {
	app := fiber.New()

	app.
		Use(favicon.New(favicon.Config{File: "public/favicon.ico"})).
		Use(requestid.New()).
		Use(logger.New(logger.Config{
			TimeFormat: "2006/01/02 15:04:05",
			Format:     "${time} | ${locals:requestid} | ${status} | ${latency} | ${reqHeader:X-Real-IP} ${ip} | ${method} | ${path} | ${error}\n",
			Next: func(c *fiber.Ctx) bool {
				return !config.I().Logging.LogHttp
			},
		})).
		Use(recover2.New(recover2.Config{
			EnableStackTrace: config.I().Logging.Level <= slog.LevelDebug,
		})).
		Use(cors.New(cors.Config{
			AllowOrigins: "http://localhost:4200",
		})).
		Use(compress.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	router := app.Group(baseUrl)
	api.Setup(router)

	app.Static(baseUrl, "./public", fiber.Static{
		Compress: true,
		MaxAge:   60 * 60,
	})
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		// This is very much nonsense, definitely have to find a better way later
		if err != nil && strings.HasPrefix(err.Error(), "Cannot GET") {
			return c.SendFile("./public/index.html")
		}

		return err
	})

	return app
}

func UpdateBaseUrlInIndex(baseUrl string) {
	if os.Getenv("DEV") != "" {
		log.Debug("Skipping base url update in DEV environment")
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get current working directory", err)
	}
	indexHtmlPath := filepath.Join(cwd, "public", "index.html")

	file, err := os.Open(indexHtmlPath)
	if err != nil {
		log.Fatal("Error opening file", err)
	}
	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Warn("Error closing file", "err", fmt.Sprintf("%+v", err))
		}
	}(file)

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		log.Fatal("Error loading HTML document", err)
	}

	doc.Find("head base").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("href", baseUrl)
	})

	html, err := doc.Html()
	if err != nil {
		log.Fatal("Error converting document to HTML", err)
	}

	err = os.WriteFile(indexHtmlPath, []byte(html), 0644)
	if err != nil {
		// Ignore errors when running as non-root in docker
		if os.Getenv("DOCKER") != "true" || !errors.Is(err, os.ErrPermission) {
			log.Fatal("Error saving modified HTML", err)
		}
	} else {
		log.Info("Updated base URL in index.html", "baseURL", baseUrl)
	}
}
