package db

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/db/repository"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setDbDriver(ctx context.Context, db *gorm.DB) error {
	settings := repository.NewSettingsRepository(db, nil)
	all, err := settings.GetAll(ctx)
	if err != nil {
		return err
	}

	all = utils.Map(all, func(setting models.ServerSetting) models.ServerSetting {
		if setting.Key != models.DbDriver {
			return setting
		}

		setting.Value = strings.ToLower(config.DbProvider)
		return setting
	})

	return settings.Update(ctx, all)
}

func migrateDrivers(ctx context.Context, log zerolog.Logger, db *gorm.DB) error {
	log = log.With().Str("handler", "db-driver-migrations").Logger()

	settings := repository.NewSettingsRepository(db, nil)
	all, err := settings.GetAll(ctx)
	if err != nil {
		return err
	}

	currentDriver := utils.Find(all, func(setting models.ServerSetting) bool {
		return setting.Key == models.DbDriver
	})

	// Not updating drivers
	if currentDriver != nil && currentDriver.Value == strings.ToLower(config.DbProvider) {
		return nil
	}

	postgresDsn := config.DatabaseDsn
	sqliteDsn := path.Join(config.Dir, "media-provider.db")

	oldDialect := func() gorm.Dialector {
		switch strings.ToLower(config.DbProvider) {
		case "sqlite":
			return postgres.Open(postgresDsn)
		case "postgres":
			return sqlite.Open(sqliteDsn)
		}
		panic(fmt.Errorf("unknown database provider: %s", config.DbProvider))
	}()

	log.WithLevel(zerolog.NoLevel).
		Str("old-driver", oldDialect.Name()).
		Str("new-driver", db.Dialector.Name()).
		Msg("Running automatic driver migrations")

	oldDb, err := gorm.Open(oldDialect, &gorm.Config{
		Logger:               gormLogger(log),
		FullSaveAssociations: true,
	})
	if err != nil {
		return err
	}
	defer func() {
		sqlDB, _ := oldDb.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	if err = cleanup(log, oldDb); err != nil {
		return fmt.Errorf("migrateDrivers cleanup old db: %w", err)
	}

	for _, model := range models.MODELS {
		if err = migrateModel(ctx, log, oldDb, db, model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}

	return nil
}

func migrateModel(ctx context.Context, log zerolog.Logger, oldDb, newDb *gorm.DB, model interface{}) error {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	log.Info().Str("model", modelType.Name()).Msg("Starting migration")

	sliceType := reflect.SliceOf(reflect.PointerTo(modelType))
	records := reflect.New(sliceType).Interface()

	if err := oldDb.WithContext(ctx).Find(records).Error; err != nil {
		return fmt.Errorf("failed to fetch records from old database: %w", err)
	}

	recordsValue := reflect.ValueOf(records).Elem()
	recordCount := recordsValue.Len()

	if recordCount == 0 {
		log.Debug().Str("model", modelType.Name()).Msg("No records to migrate")
		return nil
	}

	log.Debug().Str("model", modelType.Name()).Int("count", recordCount).Msg("Migrating records")

	const batchSize = 100
	for i := 0; i < recordCount; i += batchSize {
		end := i + batchSize
		if end > recordCount {
			end = recordCount
		}

		batch := recordsValue.Slice(i, end).Interface()

		if err := newDb.WithContext(ctx).Create(batch).Error; err != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %w", i, end-1, err)
		}
	}

	log.Debug().Str("model", modelType.Name()).Int("count", recordCount).Msg("Migration completed")
	return nil
}
