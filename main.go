package main

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/yoitsu"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var cfg *config.Config
var baseURL string
var baseURLMap fiber.Map

func init() {
	var err error
	if cfg, err = config.Load(); err != nil {
		panic(err)
	}

	opt := &slog.HandlerOptions{
		AddSource:   cfg.Logging.Source,
		Level:       cfg.Logging.Level,
		ReplaceAttr: nil,
	}
	var h slog.Handler
	switch strings.ToUpper(cfg.Logging.Handler) {
	case "TEXT":
		h = slog.NewTextHandler(os.Stdout, opt)
	case "JSON":
		h = slog.NewJSONHandler(os.Stdout, opt)
	default:
		panic("Invalid logging handler: " + cfg.Logging.Handler)
	}
	_log := slog.New(h)
	slog.SetDefault(_log)
	log.SetDefault(_log)

	validateConfig()

	baseURL = config.OrDefault(cfg.BaseUrl, "")
	baseURLMap = fiber.Map{
		"path": baseURL,
	}
	auth.Init(cfg)
	yoitsu.Init(cfg)
	mangadex.Init(cfg)
}

func main() {
	log.Info("Starting Media-Provider", "baseURL", baseURL)
	app := fiber.New()

	app.Use(logger.New(logger.Config{
		TimeFormat: "2006/01/02 15:04:05",
		Format:     "${time} | ${status} | ${latency} | ${reqHeader:X-Real-IP} ${ip} | ${method} | ${path} | ${error}\n",
		Next: func(c *fiber.Ctx) bool {
			return !cfg.Logging.LogHttp
		},
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:4200",
	}))

	app.Static(baseURL, "./UI/Web/dist/web/browser")
	router := app.Group(baseURL)
	api.Setup(router)

	port := config.OrDefault(cfg.Port, "80")
	e := app.Listen(":" + port)
	if e != nil {
		slog.Error("Unable to start server, exiting application", "error", e)
		panic(e)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	err := app.ShutdownWithTimeout(time.Second * 30)
	if err != nil {
		log.Error("An error occurred during shutdown", "error", err)
		return
	}
}
