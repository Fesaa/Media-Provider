package main

import (
	"github.com/Fesaa/Media-Provider/api/auth"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/metadata"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.uber.org/dig"
	"runtime"
)

func main() {
	c := dig.New()

	utils.Must(c.Provide(utils.Identity(afero.Afero{Fs: afero.NewOsFs()})))

	utils.Must(c.Provide(utils.Identity(c)))
	utils.Must(c.Provide(config.Load))
	utils.Must(c.Provide(LogProvider))
	utils.Must(c.Provide(services.ValidatorProvider))
	utils.Must(c.Invoke(validateConfig))

	utils.Must(c.Provide(db.DatabaseProvider))
	utils.Must(c.Invoke(db.ModelsProvider))

	utils.Must(c.Provide(auth.NewJwtAuth, dig.Name("jwt-auth")))
	utils.Must(c.Provide(auth.NewApiKeyAuth, dig.Name("api-key-auth")))

	utils.Must(c.Provide(menou.New))
	utils.Must(c.Provide(menou.NewWithRetry, dig.Name("http-retry")))
	utils.Must(c.Provide(yoitsu.New))
	utils.Must(c.Provide(pasloe.New))
	utils.Must(c.Provide(services.TranslocoServiceProvider))
	utils.Must(c.Provide(services.MarkdownServiceProvider))
	utils.Must(c.Provide(services.ValidationServiceProvider))
	utils.Must(c.Provide(services.PageServiceProvider))
	utils.Must(c.Provide(services.ContentServiceProvider))
	utils.Must(c.Provide(services.CronServiceProvider))
	utils.Must(c.Provide(services.SubscriptionServiceProvider))
	utils.Must(c.Provide(services.SignalRServiceProvider))
	utils.Must(c.Provide(services.NotificationServiceProvider))
	utils.Must(c.Provide(services.PreferenceServiceProvider))
	utils.Must(c.Provide(services.ImageServiceProvider))
	utils.Must(c.Provide(services.CacheServiceProvider))
	utils.Must(c.Provide(services.DirectoryServiceProvider))
	utils.Must(c.Provide(services.FileServiceProvider))
	utils.Must(c.Provide(services.ArchiveServiceProvider))
	utils.Must(c.Provide(services.MetadataServiceProvider))
	utils.Must(c.Provide(services.SettingsServiceProvider))
	utils.Must(c.Provide(ApplicationProvider))

	utils.Must(c.Invoke(services.RegisterSignalREndPoint))
	utils.Must(c.Invoke(RegisterCallback))
	utils.Must(c.Invoke(providers.RegisterProviders))
	utils.Must(c.Invoke(UpdateBaseUrlInIndex))
	utils.Must(c.Invoke(UpdateInstalledVersion))
	utils.Must(c.Invoke(startApp))
}

func startApp(app *fiber.App, log zerolog.Logger, cfg *config.Config) {
	log.WithLevel(zerolog.NoLevel).Str("handler", "core").
		Str("Version", metadata.Version.String()).
		Str("CommitHash", metadata.CommitHash).
		Str("BuildTimestamp", metadata.BuildTimestamp).
		Str("GoVersion", runtime.Version()).
		Str("GOOS", runtime.GOOS).
		Str("GOARCH", runtime.GOARCH).
		Str("baseUrl", cfg.BaseUrl).
		Msg("Starting")

	e := app.Listen(":8080")
	if e != nil {
		log.Fatal().Err(e).Msg("Failed to start Media-Provider")
	}
}
