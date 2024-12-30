package main

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/rs/zerolog"
	"os"
	"time"
)

func LogProvider(cfg *config.Config) zerolog.Logger {
	zerolog.DurationFieldUnit = time.Second

	ctx := func() zerolog.Logger {
		switch cfg.Logging.Handler {
		case config.LogHandlerText:
			return zerolog.New(consoleWriter())
		case config.LogHandlerJSON:
			return zerolog.New(os.Stdout)
		default:
			panic("Unknown log handler: " + cfg.Logging.Handler)
		}
	}().With()

	if cfg.Logging.Source {
		ctx = ctx.Caller()
	}

	return ctx.
		Timestamp().
		Str("handler", "core").
		Logger().
		Level(cfg.Logging.Level)
}

func consoleWriter() zerolog.ConsoleWriter {
	cw := zerolog.NewConsoleWriter()
	cw.TimeFormat = "2006-01-02 15:04:05"
	cw.PartsOrder = []string{
		zerolog.TimestampFieldName,
		zerolog.LevelFieldName,
		"handler",
		zerolog.CallerFieldName,
		zerolog.MessageFieldName,
	}
	cw.FieldsExclude = []string{
		"handler",
	}
	return cw
}
