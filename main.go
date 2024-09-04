package main

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/yoitsu"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

func init() {
	var err error
	if cfg, err = config.Load(); err != nil {
		panic(err)
	}

	log.Init(cfg.Logging)
	validateConfig()
	baseURL = config.OrDefault(cfg.BaseUrl, "")
	auth.Init()
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

	router := app.Group(baseURL)
	api.Setup(router)

	app.Static(baseURL, "./public", fiber.Static{
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

	port := config.OrDefault(cfg.Port, "80")
	e := app.Listen(":" + port)
	if e != nil {
		log.Fatal("Unable to start server, exiting application", e)
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
