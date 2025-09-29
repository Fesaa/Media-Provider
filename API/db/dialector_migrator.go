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
	"github.com/spf13/afero"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setDbDriver(ctx context.Context, db *gorm.DB) error {
	settings := repository.NewSettingsRepository(db)
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

func getOldDb(ctx context.Context, db *gorm.DB, fs afero.Afero, log zerolog.Logger) (*gorm.DB, bool, error) {
	postgresDsn := config.DatabaseDsn
	sqliteDsn := path.Join(config.Dir, "media-provider.db")

	var oldDialect gorm.Dialector
	switch strings.ToLower(config.DbProvider) {
	case "sqlite":
		if postgresDsn == "" {
			return nil, false, nil
		}
		oldDialect = postgres.Open(postgresDsn)
	case "postgres":
		ok, err := fs.Exists(sqliteDsn)
		if err != nil || !ok {
			return nil, false, err
		}

		oldDialect = sqlite.Open(sqliteDsn)
	}

	log.WithLevel(zerolog.NoLevel).
		Str("old-driver", oldDialect.Name()).
		Str("new-driver", db.Dialector.Name()).
		Msg("Running automatic driver migrations")

	oldDb, err := gorm.Open(oldDialect, &gorm.Config{
		Logger:               gormLogger(log),
		FullSaveAssociations: true,
	})
	if err != nil {
		return nil, false, err
	}

	return oldDb, true, nil
}

func migrateDrivers(ctx context.Context, log zerolog.Logger, db *gorm.DB, fs afero.Afero) error {
	log = log.With().Str("handler", "db-driver-migrations").Logger()
	if db.Migrator().HasTable("server_settings") {
		return nil
	}

	oldDb, ok, err := getOldDb(ctx, db, fs, log)
	if err != nil || !ok {
		return err
	}

	defer func() {
		sqlDB, _ := oldDb.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Ensure old db is up to date
	if err = migrate(ctx, oldDb, log); err != nil {
		return fmt.Errorf("failed to migrate old database: %w", err)
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

	// Ensure tables exist in new db
	if err := newDb.Migrator().AutoMigrate(model); err != nil {
		return fmt.Errorf("failed to migrate model in new db: %w", err)
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

		slice := recordsValue.Slice(i, end)

		for j := range slice.Len() {
			record := slice.Index(j).Addr().Interface()
			val := reflect.ValueOf(record).Elem()
			idField := val.FieldByName("ID")
			if idField.IsValid() && idField.CanSet() && idField.Kind() == reflect.Int {
				idField.SetInt(0)
			}
		}

		batch := slice.Interface()
		if err := newDb.WithContext(ctx).Create(batch).Error; err != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %w", i, end-1, err)
		}
	}

	log.Debug().Str("model", modelType.Name()).Int("count", recordCount).Msg("Migration completed")
	return nil
}
