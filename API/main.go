package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/internal/metadata"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/providers"
	"github.com/Fesaa/Media-Provider/providers/pasloe"
	"github.com/Fesaa/Media-Provider/providers/pasloe/publication"
	"github.com/Fesaa/Media-Provider/providers/yoitsu"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/dig"

	_ "net/http/pprof" //nolint: gosec
)

func main() {
	c := dig.New()

	utils.Must(c.Provide(utils.Identity(afero.Afero{Fs: afero.NewOsFs()})))

	utils.Must(c.Provide(utils.Identity(c)))
	utils.Must(c.Provide(config.Load))
	utils.Must(c.Provide(LogProvider))
	utils.Must(c.Provide(services.ValidatorProvider))
	utils.Must(c.Invoke(validateConfig))
	utils.Must(c.Invoke(setupOtel))

	ctx := context.Background()
	// span.End is called in startApp
	ctx, _ = tracing.TracerMain.Start(ctx, tracing.SpanApplicationStart) //nolint: spancheck
	utils.Must(c.Provide(utils.Identity(ctx)))

	utils.Must(c.Provide(db.DatabaseProvider))
	utils.Must(c.Provide(db.NewUnitOfWork))

	utils.Must(c.Provide(services.CookieAuthServiceProvider))
	utils.Must(c.Provide(services.ApiKeyAuthServiceProvider))

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
	utils.Must(c.Provide(services.ImageServiceProvider))
	utils.Must(c.Provide(services.CacheServiceProvider))
	utils.Must(c.Provide(services.DirectoryServiceProvider))
	utils.Must(c.Provide(services.ArchiveServiceProvider))
	utils.Must(c.Provide(services.SettingsServiceProvider))
	utils.Must(c.Provide(services.UserServiceProvider))
	utils.Must(c.Provide(applicationProvider))

	utils.Must(c.Invoke(services.RegisterSignalREndPoint))
	utils.Must(c.Invoke(registerCallback))
	utils.Must(c.Invoke(providers.RegisterProviders))
	utils.Must(c.Invoke(updateBaseUrlInIndex))
	utils.Must(c.Invoke(updateInstalledVersion))

	utils.Must(c.Invoke(startApp))
}

func startApp(c *dig.Container, app *fiber.App, log zerolog.Logger, cfg *config.Config, ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	span.End()

	log.WithLevel(zerolog.NoLevel).Str("handler", "core").
		Str("Version", metadata.Version.String()).
		Str("CommitHash", metadata.CommitHash).
		Str("BuildTimestamp", metadata.BuildTimestamp).
		Str("GoVersion", runtime.Version()).
		Str("GOOS", runtime.GOOS).
		Str("GOARCH", runtime.GOARCH).
		Str("baseUrl", cfg.BaseUrl).
		Msg("Starting")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Fatal().Str("handler", "core").Err(err).Msg("Failed to start Media-Provider")
		}
	}()

	if config.EnablePprof {
		go func() {
			log.Warn().Str("handler", "core").
				Str("adrr", "::6060").
				Msg("pprof is being registered as a handler, ensure your application is secured sufficiently. Private information may leak")
			if err := http.ListenAndServe(":6060", nil); err != nil { //nolint: gosec
				log.Fatal().Str("handler", "core").Err(err).Msg("Failed to start pprof")
			}
		}()
	}

	<-quit
	utils.Must(c.Invoke(graceFullShutdown))
}

func graceFullShutdown(app *fiber.App, log zerolog.Logger, pasloe publication.Client, yoitsu yoitsu.Client) {
	log.Info().Str("handler", "core").
		Msg("Shutting down gracefully, giving services 1 minute to shut down nicely")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if otelShutDown != nil {
		if err := otelShutDown(ctx); err != nil {
			log.Error().Str("handler", "core").Err(err).Msg("Failed to shut down Open Telemetry")
		}
	}

	var wg sync.WaitGroup

	utils.Defer(app.Shutdown, log, &wg)
	utils.Defer(pasloe.Shutdown, log, &wg)
	utils.Defer(yoitsu.Shutdown, log, &wg)

	select {
	case <-utils.Wait(&wg):
		log.Info().Str("handler", "core").
			Msg("Media Provider has shutdown nicely, Good Bye!")
	case <-ctx.Done():
		log.Warn().Str("handler", "core").
			Msg("Shutdown timed out, some content may be in a bad state")
	}
}
