package main

import (
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/frontend"
	"github.com/Fesaa/Media-Provider/impl"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	_ "github.com/lib/pq"
)

var holder models.Holder

func main() {
	engine := html.New("./web/views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
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
		},
	})

	h, err := impl.New()
	if err != nil {
		slog.Error("Cannot create holder")
		panic(err)
	}
	holder = h

	app.Use(setHolder)
	app.Hooks().OnShutdown(holder.Shutdown)
	app.Static("/", "./web/public")

	err = api.Setup(app, holder)
	if err != nil {
		slog.Error("Cannot setup api")
		panic(err)
	}

	err = frontend.Register(app)
	if err != nil {
		slog.Error("Cannot register frontend")
		panic(err)
	}

	e := app.Listen(os.Getenv("PORT"))
	if e != nil {
		slog.Error("Cannot start server")
		panic(e)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	app.ShutdownWithTimeout(time.Second * 30)
}

func setHolder(ctx *fiber.Ctx) error {
	ctx.Locals(models.HolderKey, holder)
	return ctx.Next()
}
