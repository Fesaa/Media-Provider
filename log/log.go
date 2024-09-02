package log

import (
	"context"
	"github.com/Fesaa/Media-Provider/config"
	"log/slog"
	"runtime"
	"time"
)

const (
	LevelTrace slog.Level = -8
	LevelFatal slog.Level = 12
)

var source bool

func Init(cfg *config.Logging) {
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
	Default().Error(msg, args...)
}

func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

func Trace(msg string, args ...any) {
	Default().Trace(msg, args...)
}

func Fatal(msg string, err error, args ...any) {
	Default().Fatal(msg, err, args...)
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
	if !slog.Default().Enabled(ctx, level) {
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
