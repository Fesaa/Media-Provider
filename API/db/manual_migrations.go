package db

import (
	"context"
	"errors"
	"slices"

	"github.com/Fesaa/Media-Provider/db/manual"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type migration struct {
	name string
	f    func(ctx context.Context, db *gorm.DB, log zerolog.Logger) error
}

var manualMigrations = []migration{
	{
		name: "20250403_SeedInitialMetadata",
		f:    manual.InitialMetadata,
	},
	{
		name: "20250112_SubscriptionDurationChanges",
		f:    manual.SubscriptionDurationChanges,
	},
	/*{
		name: "20250112_InsertDefaultPreferences",
		f:    manual.InsertDefaultPreferences,
	},
	{
		name: "20250215_MigrateTags",
		f:    manual.MigrateTags,
	},*/
	{
		name: "20250326_SubscriptionNextExec",
		f:    manual.SubscriptionNextExec,
	},
	{
		name: "20250327_RemoveAllDeleted",
		f:    manual.RemoveAllDeleted,
	},
	{
		name: "20250329_MigrateSettings",
		f:    manual.MigrateSettings,
	},
	{
		name: "20250903_UserRoles",
		f:    manual.UpdateUserRoles,
	},
	{
		name: "20250904_SubscriptionUserOwned",
		f:    manual.AssignSubscriptionsToUser,
	},
	{
		name: "20250905_NotificationsUserOwned",
		f:    manual.AssignNotificationsToDefaultUser,
	},
	{
		name: "20250906_RemoveRedirectUrl",
		f:    manual.RemoveRedirectUrl,
	},
	{
		name: "20250907_MigrateMetadata",
		f:    manual.MigrateMetadataToSettings,
	},
	{
		name: "20250918_PageCompactor",
		f:    manual.PageCompactor,
	},
	{
		name: "20250918_PreferenceCompactor",
		f:    manual.PreferenceCompactor,
	},
	{
		name: "20250918_AssignPreferencesToDefaultUser",
		f:    manual.AssignPreferencesToFirstAdmin,
	},
	{
		name: "20250918_SeedUserPreferences",
		f:    manual.SetDefaultUserPreferences,
	},
}

func manualMigration(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	ctx, span := tracing.TracerDb.Start(ctx, tracing.SpanManualMigrations)
	defer span.End()

	var migrations []models.ManualMigration
	if err := db.WithContext(ctx).Find(&migrations).Error; err != nil {
		return err
	}

	success := utils.MaybeMap(migrations, func(t models.ManualMigration) (string, bool) {
		if t.Success {
			return t.Name, true
		}

		log.Trace().Str("name", t.Name).Msg("Migration has ran, but was unsuccessful, trying again")
		return "", false
	})

	toDo := utils.Filter(manualMigrations, func(m migration) bool {
		return !slices.Contains(success, m.name)
	})

	log.Debug().Int("total", len(migrations)).Int("todo", len(toDo)).Msg("migrations to run")

	for _, m := range toDo {
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

			migrationLogger := log.With().Str("migration", m.name).Logger()

			migrationLogger.WithLevel(zerolog.FatalLevel).Msg("Running manual migration. This is not an error")
			var errorTally []error

			err := m.f(ctx, tx, migrationLogger)
			if err != nil {
				errorTally = append(errorTally, err)
			}

			model := models.ManualMigration{
				Name:    m.name,
				Success: len(errorTally) == 0,
			}

			err = tx.Save(&model).Error
			if err != nil {
				migrationLogger.Warn().Err(err).Msg("Failed to save manual migration")
				errorTally = append(errorTally, err)
			}

			if len(errorTally) > 0 {
				err = tx.Rollback().Error
				if err != nil {
					migrationLogger.Warn().Err(err).Msg("Failed to rollback manual migration")
				}

				errorTally = append(errorTally, err)
				return errors.Join(errorTally...)
			}

			migrationLogger.WithLevel(zerolog.FatalLevel).Msg("Finished running manual migration. This is not an error")
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
