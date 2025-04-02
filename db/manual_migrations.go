package db

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db/manual"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"slices"
)

type migration struct {
	name string
	f    func(db *gorm.DB) error
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
	{
		name: "20250112_InsertDefaultPreferences",
		f:    manual.InsertDefaultPreferences,
	},
	{
		name: "20250215_MigrateTags",
		f:    manual.MigrateTags,
	},
	{
		name: "20250326_SubscriptionNextExec",
		f:    manual.SubscriptionNextExec,
	},
	{
		name: "20250327_RemoveAllDeleted",
		f:    manual.RemoveAllDeleted,
	},
}

// TODO: Add versioning, so we don't run migrations when not needed
func manualMigration(db *gorm.DB, log zerolog.Logger) error {
	var migrations []models.ManualMigration
	if err := db.Find(&migrations).Error; err != nil {
		return err
	}

	success := utils.MaybeMap(migrations, func(t models.ManualMigration) (string, bool) {
		if t.Success {
			return t.Name, true
		}

		return "", false
	})

	toDo := utils.Filter(manualMigrations, func(m migration) bool {
		return !slices.Contains(success, m.name)
	})

	for _, m := range toDo {
		err := db.Transaction(func(tx *gorm.DB) error {
			log.WithLevel(zerolog.FatalLevel).Str("name", m.name).Msg("Running manual migration. This is not an error")
			var errorTally []error
			err := m.f(tx)
			if err != nil {
				errorTally = append(errorTally, err)
			}

			model := models.ManualMigration{
				Name:    m.name,
				Success: len(errorTally) == 0,
			}

			err = tx.Save(&model).Error
			if err != nil {
				log.Warn().Err(err).Str("name", m.name).Msg("Failed to save manual migration")
				errorTally = append(errorTally, err)
			}

			if len(errorTally) > 0 {
				err = tx.Rollback().Error
				if err != nil {
					log.Warn().Err(err).Str("name", m.name).Msg("Failed to rollback manual migration")
				}

				errorTally = append(errorTally, err)
				return errors.Join(errorTally...)
			}

			log.WithLevel(zerolog.FatalLevel).Str("name", m.name).Msg("Finished running manual migration. This is not an error")
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
