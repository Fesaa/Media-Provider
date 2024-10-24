package main

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/wisewolf"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers/mangadex"
	"github.com/Fesaa/Media-Provider/providers/webtoon"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/config"
)

var cfg *config.Config
var database *db.Database

func init() {
	var err error
	if cfg, err = config.Load(); err != nil {
		panic(err)
	}

	log.Init(cfg.Logging)
	validateConfig(cfg)
	wisewolf.Init()
	database, err = db.Connect()
	if err != nil {
		panic(err)
	}
	if err = models.Init(database.DB()); err != nil {
		log.Fatal("failed to initialize prepared statements", err)
	}

	UpdateBaseUrlInIndex(cfg.BaseUrl)
	auth.Init(database)
	yoitsu.Init(cfg)
	mangadex.Init(cfg)
	webtoon.Init(cfg)
}

func main() {
	defer db.Close()
	log.Info("Starting Media-Provider", "baseURL", cfg.BaseUrl)

	app := SetupApp(cfg.BaseUrl)

	e := app.Listen(":8080")
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
