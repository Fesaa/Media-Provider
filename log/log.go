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

func Error(msg string, args ...any) {
	log(slog.LevelError, msg, args...)
}

func Warn(msg string, args ...any) {
	log(slog.LevelWarn, msg, args...)
}

func Info(msg string, args ...any) {
	log(slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	log(slog.LevelDebug, msg, args...)
}

func Trace(msg string, args ...any) {
	log(LevelTrace, msg, args...)
}

func Fatal(msg string, args ...any) {
	log(LevelFatal, msg, args...)
	panic("fatal log call")
}

// Overriding default, such that the source is correct
// this is almost a copy of the slog.log function
func log(level slog.Level, msg string, args ...any) {
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

	_ = slog.Default().Handler().Handle(ctx, r)
}
