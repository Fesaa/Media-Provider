package main

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/rs/zerolog"
	"os"
)

func LogProvider(cfg *config.Config) zerolog.Logger {
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
		Logger().
		Level(zerolog.Level(cfg.Logging.Level))
}

func consoleWriter() zerolog.ConsoleWriter {
	cw := zerolog.NewConsoleWriter()
	cw.TimeFormat = "2006-01-02 15:04:05"
	return cw
}
