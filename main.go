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
	"strings"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

var baseURL string
var baseURLMap fiber.Map

func init() {
	if err := config.LoadConfig("config.yaml"); err != nil {
		panic(err)
	}

	opt := &slog.HandlerOptions{
		AddSource:   config.I().GetLoggingConfig().GetSource(),
		Level:       config.I().GetLoggingConfig().GetLogLevel(),
		ReplaceAttr: nil,
	}
	var h slog.Handler
	switch strings.ToUpper(config.I().GetLoggingConfig().GetHandler()) {
	case "TEXT":
		h = slog.NewTextHandler(os.Stdout, opt)
	case "JSON":
		h = slog.NewJSONHandler(os.Stdout, opt)
	default:
		panic("Invalid logging handler: " + config.I().GetLoggingConfig().GetHandler())
	}
	slog.SetDefault(slog.New(h))
	validateConfig()

	baseURL = config.OrDefault(config.I().GetRootURl(), "")
	baseURLMap = fiber.Map{
		"path": baseURL,
	}
	auth.Init()
	yoitsu.Init(config.I())
	mangadex.Init(config.I())
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
			return !config.I().GetLoggingConfig().LogHttp()
		},
	}))

	app.Static(baseURL, "./web/public")
	router := app.Group(baseURL)
	api.Setup(router)
	RegisterFrontEnd(router)

	port := config.OrDefault(config.I().GetPort(), "80")
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
