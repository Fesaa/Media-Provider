package main

import (
	"errors"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/yoitsu"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

var cfg *config.Config
var baseURL string
var baseURLMap fiber.Map

func init() {
	var err error

	file := config.OrDefault(os.Getenv("CONFIG_FILE"), "config.json")
	if cfg, err = config.Load(path.Join("config", file)); err != nil {
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
	engine := html.New("./web/views", ".html")
	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: errorHandler,
	})

	app.Use(logger.New(logger.Config{
		TimeFormat: "2006/01/02 15:04:05",
		Format:     "${time} | ${status} | ${latency} | ${reqHeader:X-Real-IP} ${ip} | ${method} | ${path} | ${error}\n",
		Next: func(c *fiber.Ctx) bool {
			return !cfg.Logging.LogHttp
		},
	}))

	app.Static(baseURL, "./web/public")
	router := app.Group(baseURL)
	router.Use(func(c *fiber.Ctx) error {
		c.Locals("cfg", cfg)
		return c.Next()
	})
	api.Setup(router)
	RegisterFrontEnd(router)

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

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	if code == fiber.StatusNotFound {
		return c.Render("404", nil)
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	return c.Status(code).SendString(err.Error())
}
