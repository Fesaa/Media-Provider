package main

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/impl"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/Fesaa/Media-Provider/mount"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

var holder models.Holder
var baseURL string
var baseURLMap fiber.Map

// Following env variables are required:
// USER, DOMAIN, URL, PASS
//
// The following env variables are optional:
//
// PORT: 80
//
// PASSWORD: admin
//
// TORRENT_DIR: temp
//
// BASE_URL: /
func main() {
	mount.Init()
	mount.Mount(true)
	baseURL = utils.GetEnv("BASE_URL", "")
	baseURLMap = fiber.Map{
		"path": baseURL,
	}

	engine := html.New("./web/views", ".html")
	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: errorHandler,
	})

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

	port := utils.GetEnv("PORT", "80")
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
