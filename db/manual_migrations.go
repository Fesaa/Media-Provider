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
		name: "20250112_SubscriptionDurationChanges",
		f:    manual.SubscriptionDurationChanges,
	},
	{
		name: "20250112_InsertDefaultPreferences",
		f:    manual.InsertDefaultPreferences,
	},
	{
		name: "20250215_InsertEmptyArray",
		f:    manual.InsertEmptyArray,
	},
	{
		name: "20250215_InsertEmptyBlackList",
		f:    manual.InsertEmptyBlackList,
	},
}

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
			log.Info().Str("name", m.name).Msg("Running manual migration")
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

			log.Info().Str("name", m.name).Msg("Finished running manual migration")
			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}
