package db

import (
	"context"
	"fmt"
	"path"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/manual"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/metadata"
	mptracing "github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func DatabaseProvider(ctx context.Context, log zerolog.Logger) (*gorm.DB, error) {
	ctx, span := mptracing.TracerDb.Start(ctx, mptracing.SpanSetupDb)
	defer span.End()

	dsn := utils.OrElse(config.DatabaseDsn, path.Join(config.Dir, "media-provider.db"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
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
		tracing.WithDBSystem("sqlite"),
	))
	if err != nil {
		return nil, err
	}

	if err = migrate(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to migrate db: %w", err)
	}

	if err = manualMigration(ctx, db, log.With().Str("handler", "migrations").Logger()); err != nil {
		return nil, fmt.Errorf("failed manual migrations: %w", err)
	}

	if err = manual.SyncSettings(db.WithContext(ctx), log.With().Str("handler", "settings").Logger()); err != nil {
		return nil, fmt.Errorf("failed to sync settings: %w", err)
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
