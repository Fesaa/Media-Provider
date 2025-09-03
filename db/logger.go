package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var defaultConfig = logger.Config{
	SlowThreshold:             1 * time.Second,
	LogLevel:                  logger.Warn,
	IgnoreRecordNotFoundError: true,
	Colorful:                  true,
}

type gormLoggerWrapper struct {
	log    zerolog.Logger
	config logger.Config
}

func levelConv(lvl logger.LogLevel) zerolog.Level {
	switch lvl {
	case logger.Silent:
		return zerolog.Disabled
	case logger.Error:
		return zerolog.ErrorLevel
	case logger.Warn:
		return zerolog.WarnLevel
	case logger.Info:
		return zerolog.InfoLevel
	default:
		return zerolog.ErrorLevel
	}
}

func gormLogger(log zerolog.Logger, config ...logger.Config) logger.Interface {
	return &gormLoggerWrapper{
		log: log.With().Str("handler", "db").Logger(),
		config: func() logger.Config {
			if len(config) > 0 {
				return config[0]
			}
			return defaultConfig
		}(),
	}
}

func (g *gormLoggerWrapper) LogMode(level logger.LogLevel) logger.Interface {
	*g = gormLoggerWrapper{
		log: g.log.Level(levelConv(level)),
	}
	return g
}

func (g *gormLoggerWrapper) Info(ctx context.Context, s string, i ...interface{}) {
	g.log.Info().Msgf(s, i...)
}

func (g *gormLoggerWrapper) Warn(ctx context.Context, s string, i ...interface{}) {
	g.log.Warn().Msgf(s, i...)
}

func (g *gormLoggerWrapper) Error(ctx context.Context, s string, i ...interface{}) {
	g.log.Error().Msgf(s, i...)
}

func (g *gormLoggerWrapper) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.log.GetLevel() >= zerolog.Disabled {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && g.log.GetLevel() <= zerolog.ErrorLevel && (!errors.Is(err, gorm.ErrRecordNotFound) || !g.config.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			g.log.Error().Err(err).Dur("elapsed", elapsed).Str("sql", sql).Send()
		} else {
			g.log.Error().Err(err).Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql).Send()
		}
	case elapsed > g.config.SlowThreshold && g.config.SlowThreshold != 0 && g.log.GetLevel() <= zerolog.WarnLevel:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", g.config.SlowThreshold)
		if rows == -1 {
			g.log.Warn().Dur("elapsed", elapsed).Str("sql", sql).Msg(slowLog)
		} else {
			g.log.Warn().Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql).Msg(slowLog)
		}
	case g.log.GetLevel() <= zerolog.TraceLevel:
		sql, rows := fc()
		if rows == -1 {
			g.log.Trace().Dur("elapsed", elapsed).Str("sql", sql).Send()
		} else {
			g.log.Trace().Dur("elapsed", elapsed).Int64("rows", rows).Str("sql", sql).Send()
		}
	}
}
