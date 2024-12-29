package main

import (
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/wisewolf"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/subscriptions"
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

	if !cfg.HasUpdatedDB {
		fmt.Println("Please migrate your database using the python script, then manually edit the config value")
		fmt.Println("Your application is not crashing, this is intended. ")
		os.Exit(1)
	}

	log.Init(cfg.Logging)
	validateConfig(cfg)
	wisewolf.Init()
	database, err = db.Connect()
	if err != nil {
		panic(err)
	}

	UpdateBaseUrlInIndex(cfg.BaseUrl)
	auth.Init(database)
	yoitsu.Init(cfg)
	pasloe.Init(cfg)
	subscriptions.Init(database)
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
