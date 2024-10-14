package main

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/mangadex"
	"github.com/Fesaa/Media-Provider/wisewolf"
	"github.com/Fesaa/Media-Provider/yoitsu"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/config"
)

var cfg *config.Config

func init() {
	var err error
	if cfg, err = config.Load(); err != nil {
		panic(err)
	}

	log.Init(cfg.Logging)
	validateConfig(cfg)
	wisewolf.Init()

	UpdateBaseUrlInIndex(cfg.BaseUrl)
	auth.Init()
	yoitsu.Init(cfg)
	mangadex.Init(cfg)
}

func main() {
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
