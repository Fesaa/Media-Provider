package main

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/metadata"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"gopkg.in/natefinch/lumberjack.v2"
)

var otelShutDown func(context.Context) error

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

func setupOtel(log zerolog.Logger) error {
	ctx := context.Background()
	log = log.With().Str("handler", "otel").Logger()

	if config.OtelEndpoint == "" {
		log.Debug().Msg("No external Otel endpoint configured")
		return nil
	}

	log.Debug().Str("endpoint", config.OtelEndpoint).Msg("Setting up Open Telemetry")

	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(config.OtelEndpoint),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	}

	if config.Development && !strings.HasPrefix(config.OtelEndpoint, "https") {
		options = append(options, otlptracehttp.WithInsecure())
	}

	if config.OtelAuth != "" {
		options = append(options, otlptracehttp.WithHeaders(map[string]string{
			"Authorization": config.OtelAuth,
		}))
	}

	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return err
	}

	res, err := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(metadata.Identifier)))
	if err != nil {
		return err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otelShutDown = tp.Shutdown
	return nil
}
