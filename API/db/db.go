package db

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/manual"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/metadata"
	mptracing "github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func DatabaseProvider(ctx context.Context, log zerolog.Logger, fs afero.Afero) (*gorm.DB, error) {
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

	err = db.Transaction(func(tx *gorm.DB) error {
		if err = migrateDrivers(ctx, log, tx, fs); err != nil {
			return fmt.Errorf("failed to migrate drivers: %w", err)
		}

		if err = migrate(ctx, tx, log); err != nil {
			return fmt.Errorf("failed to migrate: %w", err)
		}

		if err = setDbDriver(ctx, tx); err != nil {
			return fmt.Errorf("failed to set db driver: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	ctx, span := mptracing.TracerDb.Start(ctx, mptracing.SpanMigrations)
	defer span.End()

	if err := db.WithContext(ctx).AutoMigrate(models.MODELS...); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	if err := manualMigration(ctx, db, log.With().Str("handler", "migrations").Logger()); err != nil {
		return fmt.Errorf("failed manual migrations: %w", err)
	}

	if err := manual.SyncSettings(db.WithContext(ctx), log.With().Str("handler", "settings").Logger()); err != nil {
		return fmt.Errorf("failed to sync settings: %w", err)
	}

	return nil
}

func getDialector(log zerolog.Logger) gorm.Dialector {
	log.Debug().Str("dialect", config.DbProvider).Msg("Using dialect")
	switch strings.ToLower(config.DbProvider) {
	case "sqlite":
		return sqlite.Open(path.Join(config.Dir, "media-provider.db"))
	case "postgres":
		return postgres.Open(config.DatabaseDsn)
	}

	panic(fmt.Errorf("unknown database provider: %s", config.DbProvider))
}
