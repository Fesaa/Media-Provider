package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/api"
	"github.com/Fesaa/Media-Provider/frontend"
	"github.com/Fesaa/Media-Provider/impl"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/sqlite3/v2"
	"github.com/gofiber/template/html/v2"
)

var holder models.Holder

func main() {
	engine := html.New("./web/views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	h, err := createHolder()
	if err != nil {
		panic(err)
	}
	holder = h

	app.Use(setHolder)
	app.Hooks().OnShutdown(holder.Shutdown)
	app.Static("/", "./web/public")

	err = api.Setup(app, holder)
	if err != nil {
		panic(err)
	}

	err = frontend.Register(app)
	if err != nil {
		panic(err)
	}

	e := app.Listen(":3000")
	if e != nil {
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

func createHolder() (models.Holder, error) {
	s := sqlite3.New(sqlite3.Config{
		Database: "./mp-db.db",
		Table:    "storageprovider",
	})
	holder, err := impl.New(s, s.Conn())
	if err != nil {
		return nil, err
	}
	return holder, nil
}
