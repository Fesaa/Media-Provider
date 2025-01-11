package db

import (
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
		name: "SubscriptionDurationChanges",
		f:    subscriptionDurationChanges,
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
		log.Info().Str("name", m.name).Msg("Running manual migration")
		if err := m.f(db); err != nil {
			log.Error().Err(err).Str("name", m.name).Msg("Failed to run migration")
			return err
		}

		model := models.ManualMigration{
			Name:    m.name,
			Success: true,
		}

		if err := db.Save(&model).Error; err != nil {
			log.Warn().Err(err).Str("name", m.name).Msg("Failed to save manual migration")
			return err
		}
		log.Info().Str("name", m.name).Msg("Finished running manual migration")
	}

	return nil
}

func subscriptionDurationChanges(db *gorm.DB) error {
	var subscriptions []models.Subscription
	res := db.Find(&subscriptions)
	if res.Error != nil {
		return res.Error
	}

	for _, sub := range subscriptions {
		if sub.RefreshFrequency < 2 {
			sub.RefreshFrequency = 2
			if err := db.Save(&sub).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
