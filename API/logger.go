package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func LogProvider(cfg *config.Config) zerolog.Logger {
	zerolog.DurationFieldUnit = time.Second
	zerolog.CallerMarshalFunc = callerMarshal
	zerolog.FormattedLevels = map[zerolog.Level]string{
		zerolog.TraceLevel: "TRACE",
		zerolog.DebugLevel: "DEBUG",
		zerolog.InfoLevel:  "INFO",
		zerolog.WarnLevel:  "WARN",
		zerolog.ErrorLevel: "ERROR",
		zerolog.FatalLevel: "FATAL",
		zerolog.PanicLevel: "PANIC",
	}

	writer := func() io.Writer {
		switch cfg.Logging.Handler {
		case config.LogHandlerText:
			return consoleWriter()
		case config.LogHandlerJSON:
			return io.MultiWriter(os.Stdout, fileWriter())
		default:
			panic("Unknown log handler: " + cfg.Logging.Handler)
		}
	}()

	ctx := zerolog.New(writer).With()

	if cfg.Logging.Source {
		ctx = ctx.Caller()
	}

	// Application logger logs everything, zerolog.GlobalLevel is only source of truth.
	// Allows for updating the log level without restarting the application
	zerolog.SetGlobalLevel(cfg.Logging.Level)
	return ctx.
		Timestamp().
		Logger().
		Level(zerolog.TraceLevel)
}

func callerMarshal(pc uintptr, file string, line int) string {
	if strings.Contains(file, "go/pkg/mod") {
		file = filepath.Base(file)
	}
	return file + ":" + strconv.Itoa(line)
}

func fileWriter() io.Writer {
	return &lumberjack.Logger{
		Filename:  path.Join(config.Dir, "logs", "media-provider.log"),
		MaxAge:    30,
		LocalTime: false,
		Compress:  true,
	}
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

	cw.NoColor = true
	cw.Out = io.MultiWriter(os.Stdout, fileWriter())
	return cw
}
