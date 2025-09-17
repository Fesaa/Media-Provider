package db

import (
	"fmt"
	"slices"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func cleanup(log zerolog.Logger, db *gorm.DB) error {
	log = log.With().Str("handler", "db-cleanup").Logger()
	log.Debug().Msg("Running db cleanup")

	return db.Transaction(func(tx *gorm.DB) error {
		// Clean up in dependency order (children first, then parents)
		if err := cleanupModifierValues(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup modifier values: %w", err)
		}

		if err := cleanupModifiers(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup modifiers: %w", err)
		}

		if err := cleanupTagMaps(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup tag maps: %w", err)
		}

		return nil
	})
}

func cleanupModifierValues(log zerolog.Logger, tx *gorm.DB) error {
	var modifierValues []models.ModifierValue
	var modifiers []models.Modifier

	if err := tx.Find(&modifierValues).Error; err != nil {
		return fmt.Errorf("failed to get modifier values: %w", err)
	}
	if err := tx.Find(&modifiers).Error; err != nil {
		return fmt.Errorf("failed to get modifiers: %w", err)
	}

	modifierIds := utils.Map(modifiers, func(modifier models.Modifier) int {
		return modifier.ID
	})

	for _, mv := range modifierValues {
		if !slices.Contains(modifierIds, mv.ModifierID) {
			log.Warn().
				Int("modifierValue_id", mv.ID).
				Int("modifier_id", mv.ModifierID).
				Msg("modifierValue does not belong to a modifier, removing")

			if err := tx.Delete(&mv).Error; err != nil {
				return fmt.Errorf("failed to delete modifier value: %w", err)
			}
		}
	}

	return nil
}

func cleanupModifiers(log zerolog.Logger, tx *gorm.DB) error {
	var modifiers []models.Modifier
	var pages []models.Page

	if err := tx.Find(&modifiers).Error; err != nil {
		return fmt.Errorf("failed to get modifiers: %w", err)
	}
	if err := tx.Find(&pages).Error; err != nil {
		return fmt.Errorf("failed to get pages: %w", err)
	}

	pageIds := utils.Map(pages, func(page models.Page) int {
		return page.ID
	})

	for _, m := range modifiers {
		if !slices.Contains(pageIds, m.PageID) {
			log.Warn().
				Int("modifier_id", m.ID).
				Int("page_id", m.PageID).
				Msg("modifier does not belong to a page, removing")

			if err := tx.Delete(&m).Error; err != nil {
				return fmt.Errorf("failed to delete modifier: %w", err)
			}
		}
	}

	return nil
}

func cleanupTagMaps(log zerolog.Logger, tx *gorm.DB) error {
	var tagMaps []models.TagMap
	if err := tx.Find(&tagMaps).Error; err != nil {
		return fmt.Errorf("failed to get tag maps: %w", err)
	}

	for _, m := range tagMaps {
		if m.PreferenceID == 0 {
			log.Warn().
				Int("tagMapID", m.ID).
				Msg("tagMap does not belong to a tag map, removing")
			if err := tx.Delete(&m).Error; err != nil {
				return fmt.Errorf("failed to delete modifier: %w", err)
			}
		}
	}
	return nil
}
