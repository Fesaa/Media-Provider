package db

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/manual"
	"github.com/Fesaa/Media-Provider/db/models"
	mptracing "github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/metadata"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func DatabaseProvider(ctx context.Context, log zerolog.Logger) (*gorm.DB, error) {
	log = log.With().Str("handler", "db").Logger()

	ctx, span := mptracing.TracerDb.Start(ctx, mptracing.SpanSetupDb)
	defer span.End()

	db, err := gorm.Open(getDialector(log), &gorm.Config{
		Logger:               gormLogger(log),
		FullSaveAssociations: true,
	})
	if err != nil {
		return nil, err
	}

	err = db.Use(tracing.NewPlugin(
		tracing.WithRecordStackTrace(),
		tracing.WithoutQueryVariables(),
		tracing.WithAttributes(semconv.ServiceName(metadata.Identifier)),
		tracing.WithDBSystem(config.DbProvider),
	))
	if err != nil {
		return nil, err
	}

	if err = migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	if err = cleanup(log, db); err != nil {
		return nil, fmt.Errorf("cleanup: %w", err)
	}

	if err = migrateDrivers(ctx, log, db); err != nil {
		return nil, fmt.Errorf("migrateDrivers: %w", err)
	}

	if err = setDbDriver(ctx, db); err != nil {
		return nil, fmt.Errorf("setDbDriver: %w", err)
	}

	if err = manualMigration(ctx, db, log.With().Str("handler", "migrations").Logger()); err != nil {
		return nil, fmt.Errorf("manualMigration: %w", err)
	}

	if err = manual.SyncSettings(db.WithContext(ctx), log.With().Str("handler", "settings").Logger()); err != nil {
		return nil, fmt.Errorf("manualSyncSettings: %w", err)
	}

	return db, nil
}

// Method for custom span
func migrate(ctx context.Context, db *gorm.DB) error {
	ctx, span := mptracing.TracerDb.Start(ctx, mptracing.SpanMigrations)
	defer span.End()

	if err := db.WithContext(ctx).AutoMigrate(models.MODELS...); err != nil {
		return err
	}

	return nil
}

func getDialector(log zerolog.Logger) gorm.Dialector {
	dsn := utils.OrElse(config.DatabaseDsn, path.Join(config.Dir, "media-provider.db"))

	log.Debug().Str("dialect", config.DbProvider).Msg("Using dialect")
	switch strings.ToLower(config.DbProvider) {
	case "sqlite":
		return sqlite.Open(dsn)
	case "postgres":
		return postgres.Open(dsn)
	}

	panic(fmt.Errorf("unknown database provider: %s", config.DbProvider))
}
