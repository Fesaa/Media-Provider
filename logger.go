package main

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func LogProvider(cfg *config.Config) zerolog.Logger {
	zerolog.DurationFieldUnit = time.Second
	zerolog.CallerMarshalFunc = callerMarshal

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

	// Application logger logs everything, zerolog.GlobalLevel is only source of truth.
	// Allows for updating the log level without restarting the application
	zerolog.SetGlobalLevel(cfg.Logging.Level)
	return ctx.
		Timestamp().
		Str("handler", "core").
		Logger().
		Level(zerolog.TraceLevel)
}

func callerMarshal(pc uintptr, file string, line int) string {
	if strings.Contains(file, "go/pkg/mod") {
		file = filepath.Base(file)
	}
	return file + ":" + strconv.Itoa(line)
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
