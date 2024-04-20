package main

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/impl"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

var holder models.Holder
var baseURL string
var baseURLMap fiber.Map

func init() {
	if err := config.LoadConfig("config.yaml"); err != nil {
		panic(err)
	}
}

func main() {
	baseURL = config.OrDefault(config.C.RootURL, "")
	baseURLMap = fiber.Map{
		"path": baseURL,
	}
	engine := html.New("./web/views", ".html")
	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: errorHandler,
	})

	app.Use(logger.New(logger.Config{
		TimeFormat: "2006/01/02 15:04:05",
		Format:     "${time} | ${status} | ${latency} | ${reqHeader:X-Real-IP} ${ip} | ${method} | ${path} | ${error}\n",
	}))

	var err error
	holder, err = impl.New()
	if err != nil {
		slog.Error("Cannot create holder")
		panic(err)
	}

	app.Use(setHolder)
	app.Hooks().OnShutdown(holder.Shutdown)
	app.Static(baseURL, "./web/public")

	router := app.Group(baseURL)

	api.Setup(router, holder)
	RegisterFrontEnd(router)

	port := config.OrDefault(config.C.Port, "80")
	e := app.Listen(":" + port)
	if e != nil {
		slog.Error("Cannot start server")
		panic(e)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	app.ShutdownWithTimeout(time.Second * 30)
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

func setHolder(ctx *fiber.Ctx) error {
	ctx.Locals(models.HolderKey, holder)
	return ctx.Next()
}
