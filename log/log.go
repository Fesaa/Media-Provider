package log

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/Fesaa/Media-Provider/config"
)

const (
	LevelTrace slog.Level = -8
	LevelFatal slog.Level = 12
)

var source bool

// Always initialize with default values, so that the logger is always usable
func init() {
	Init(config.Logging{
		Handler: config.LogHandlerText,
		Level:   slog.LevelInfo,
		Source:  false,
	})
}

func Init(cfg config.Logging) {
	opt := &slog.HandlerOptions{
		AddSource:   cfg.Source,
		Level:       cfg.Level,
		ReplaceAttr: nil,
	}
	var h slog.Handler
	switch cfg.Handler {
	case config.LogHandlerText:
		h = slog.NewTextHandler(os.Stdout, opt)
	case config.LogHandlerJSON:
		h = slog.NewJSONHandler(os.Stdout, opt)
	default:
		panic("Invalid logging handler: " + cfg.Handler)
	}
	_log := slog.New(h)
	slog.SetDefault(_log)
	SetDefault(_log)
	source = cfg.Source
	def = Logger{_log: slog.Default(), source: nil}
}

type Logger struct {
	_log   *slog.Logger
	source *bool
}

func SetDefault(log *slog.Logger) {
	def = Logger{_log: log}
}

func Default() *Logger {
	return &def
}

var def Logger

func IsTraceEnabled() bool {
	return Default().IsTraceEnabled()
}

func Error(msg string, args ...any) {
	Default().log(slog.LevelError, msg, args...)
}

func Warn(msg string, args ...any) {
	Default().log(slog.LevelWarn, msg, args...)
}

func Info(msg string, args ...any) {
	Default().log(slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	Default().log(slog.LevelDebug, msg, args...)
}

func Trace(msg string, args ...any) {
	Default().log(LevelTrace, msg, args...)
}

func Fatal(msg string, err error, args ...any) {
	Default().log(LevelFatal, msg, args...)
	panic(err)
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.log(slog.LevelWarn, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(slog.LevelDebug, msg, args...)
}

func (l *Logger) Trace(msg string, args ...any) {
	l.log(LevelTrace, msg, args...)
}

func (l *Logger) Fatal(msg string, err error, args ...any) {
	l.log(LevelFatal, msg, args...)
	panic(err)
}

func (l *Logger) IsTraceEnabled() bool {
	return l._log.Enabled(nil, LevelTrace)
}

func With(args ...any) *Logger {
	return Default().With(args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		_log: l._log.With(args...),
	}
}

func (l *Logger) SetSource(source bool) {
	l.source = &source
}

// Overriding default, such that the source is correct
// this is almost a copy of the slog.log function
func (l *Logger) log(level slog.Level, msg string, args ...any) {
	ctx := context.Background()
	if !l._log.Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if (l.source != nil && *l.source) || (source && l.source == nil) {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)

	_ = l._log.Handler().Handle(ctx, r)
}
