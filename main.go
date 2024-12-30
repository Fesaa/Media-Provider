package main

import (
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/wisewolf"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/subscriptions"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

func main() {
	c := dig.New()

	utils.Must(c.Provide(utils.Identity(c)))
	utils.Must(c.Provide(config.Load))
	utils.Must(c.Provide(LogProvider))
	utils.Must(c.Invoke(validateConfig))

	utils.Must(c.Provide(db.DatabaseProvider))
	utils.Must(c.Provide(auth.NewJwtAuth, dig.Name("jwt-auth")))
	utils.Must(c.Provide(auth.NewApiKeyAuth, dig.Name("api-key-auth")))

	utils.Must(c.Provide(wisewolf.New))
	utils.Must(c.Provide(mangadex.NewRepository))

	utils.Must(c.Provide(yoitsu.New))
	utils.Must(c.Provide(pasloe.New))
	utils.Must(c.Provide(providers.New))
	utils.Must(c.Provide(subscriptions.New))
	utils.Must(c.Provide(ApplicationProvider))

	utils.Must(c.Invoke(UpdateBaseUrlInIndex))
	utils.Must(c.Invoke(startApp))
}

func startApp(app *fiber.App, log zerolog.Logger, cfg *config.Config) {
	log.Info().Str("baseUrl", cfg.BaseUrl).Msg("Starting Media-Provider")

	e := app.Listen(":8080")
	if e != nil {
		log.Fatal().Err(e).Msg("Failed to start Media-Provider")
	}
}
