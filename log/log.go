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

func init() {
	def = Logger{_log: slog.Default()}
}

type Logger struct {
	_log *slog.Logger
}

func SetDefault(log *slog.Logger) {
	def = Logger{_log: log}
}

func Default() *Logger {
	return &def
}

var def Logger

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

func Fatal(msg string, args ...any) {
	Default().log(LevelFatal, msg, args...)
	panic("fatal log call")
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

func (l *Logger) Fatal(msg string, args ...any) {
	l.log(LevelFatal, msg, args...)
	panic("fatal log call")
}

func With(args ...any) *Logger {
	return Default().With(args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		_log: l._log.With(args...),
	}
}

// Overriding default, such that the source is correct
// this is almost a copy of the slog.log function
func (l *Logger) log(level slog.Level, msg string, args ...any) {
	ctx := context.Background()
	if !slog.Default().Enabled(ctx, level) {
		return
	}
	var pc uintptr
	if config.I().GetLoggingConfig().GetSource() {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)

	_ = l._log.Handler().Handle(ctx, r)
}
