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

		if err := cleanupAgeRatingMaps(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup age rating maps: %w", err)
		}

		if err := cleanupTags(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup tags: %w", err)
		}

		if err := cleanupTagMaps(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup tag maps: %w", err)
		}

		if err := cleanupManyToManyRelations(log, tx); err != nil {
			return fmt.Errorf("failed to cleanup many to many relations: %w", err)
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

func cleanupTags(log zerolog.Logger, tx *gorm.DB) error {
	var tags []models.Tag
	var preferences []models.Preference
	var ageRatingMaps []models.AgeRatingMap

	if err := tx.Find(&tags).Error; err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}
	if err := tx.Find(&preferences).Error; err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}
	if err := tx.Find(&ageRatingMaps).Error; err != nil {
		return fmt.Errorf("failed to get age rating maps: %w", err)
	}

	preferenceIds := utils.Map(preferences, func(pref models.Preference) int {
		return pref.ID
	})
	ageRatingMapIds := utils.Map(ageRatingMaps, func(arm models.AgeRatingMap) int {
		return arm.ID
	})

	for _, tag := range tags {
		shouldDelete := false
		reason := ""

		// Check if tag belongs to a valid preference (if PreferenceID is set and not 0)
		if tag.PreferenceID != 0 && !slices.Contains(preferenceIds, tag.PreferenceID) {
			shouldDelete = true
			reason = "preference does not exist"
		}

		// Check if tag belongs to a valid age rating map (if AgeRatingMapID is set and not 0)
		if !shouldDelete && tag.AgeRatingMapID != 0 && !slices.Contains(ageRatingMapIds, tag.AgeRatingMapID) {
			shouldDelete = true
			reason = "age rating map does not exist"
		}

		if shouldDelete {
			log.Warn().
				Int("tag_id", tag.ID).
				Int("preference_id", tag.PreferenceID).
				Int("age_rating_map_id", tag.AgeRatingMapID).
				Str("reason", reason).
				Msg("tag has invalid foreign key reference, removing")

			if err := tx.Delete(&tag).Error; err != nil {
				return fmt.Errorf("failed to delete tag: %w", err)
			}
		}
	}

	return nil
}

func cleanupAgeRatingMaps(log zerolog.Logger, tx *gorm.DB) error {
	var ageRatingMaps []models.AgeRatingMap
	var preferences []models.Preference

	if err := tx.Find(&ageRatingMaps).Error; err != nil {
		return fmt.Errorf("failed to get age rating maps: %w", err)
	}
	if err := tx.Find(&preferences).Error; err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	preferenceIds := utils.Map(preferences, func(pref models.Preference) int {
		return pref.ID
	})

	for _, arm := range ageRatingMaps {
		if !slices.Contains(preferenceIds, arm.PreferenceID) {
			log.Warn().
				Int("age_rating_map_id", arm.ID).
				Int("preference_id", arm.PreferenceID).
				Msg("age rating map does not belong to a preference, removing")

			if err := tx.Delete(&arm).Error; err != nil {
				return fmt.Errorf("failed to delete age rating map: %w", err)
			}
		}
	}

	return nil
}

func cleanupTagMaps(log zerolog.Logger, tx *gorm.DB) error {
	var tagMaps []models.TagMap
	var preferences []models.Preference
	var tags []models.Tag

	if err := tx.Find(&tagMaps).Error; err != nil {
		return fmt.Errorf("failed to get tag maps: %w", err)
	}
	if err := tx.Find(&preferences).Error; err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}
	if err := tx.Find(&tags).Error; err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}

	preferenceIds := utils.Map(preferences, func(pref models.Preference) int {
		return pref.ID
	})
	tagIds := utils.Map(tags, func(tag models.Tag) int {
		return tag.ID
	})

	for _, tm := range tagMaps {
		shouldDelete := false
		reason := ""

		if !slices.Contains(preferenceIds, tm.PreferenceID) {
			shouldDelete = true
			reason = "preference does not exist"
		} else if !slices.Contains(tagIds, tm.OriginID) {
			shouldDelete = true
			reason = "origin tag does not exist"
		} else if !slices.Contains(tagIds, tm.DestID) {
			shouldDelete = true
			reason = "destination tag does not exist"
		}

		if shouldDelete {
			log.Warn().
				Int("tag_map_id", tm.ID).
				Int("preference_id", tm.PreferenceID).
				Int("origin_id", tm.OriginID).
				Int("dest_id", tm.DestID).
				Str("reason", reason).
				Msg("tag map has invalid foreign key reference, removing")

			if err := tx.Delete(&tm).Error; err != nil {
				return fmt.Errorf("failed to delete tag map: %w", err)
			}
		}
	}

	return nil
}

// Optional: Cleanup many-to-many relationships
func cleanupManyToManyRelations(log zerolog.Logger, tx *gorm.DB) error {
	// Clean up preference_dynasty_genre_tags
	if err := cleanupManyToMany(log, tx, "preference_dynasty_genre_tags", "preference_id", "tag_id", "preferences", "tags"); err != nil {
		return fmt.Errorf("failed to cleanup preference_dynasty_genre_tags: %w", err)
	}

	// Clean up preference_black_list_tags
	if err := cleanupManyToMany(log, tx, "preference_black_list_tags", "preference_id", "tag_id", "preferences", "tags"); err != nil {
		return fmt.Errorf("failed to cleanup preference_black_list_tags: %w", err)
	}

	// Clean up preference_white_list_tags
	if err := cleanupManyToMany(log, tx, "preference_white_list_tags", "preference_id", "tag_id", "preferences", "tags"); err != nil {
		return fmt.Errorf("failed to cleanup preference_white_list_tags: %w", err)
	}

	// Clean up preference_age_rating_mappings
	if err := cleanupManyToMany(log, tx, "preference_age_rating_mappings", "preference_id", "age_rating_map_id", "preferences", "age_rating_maps"); err != nil {
		return fmt.Errorf("failed to cleanup preference_age_rating_mappings: %w", err)
	}

	return nil
}

func cleanupManyToMany(log zerolog.Logger, tx *gorm.DB, tableName, leftCol, rightCol, leftTable, rightTable string) error {
	query := fmt.Sprintf(`
        DELETE FROM %s 
        WHERE %s NOT IN (SELECT id FROM %s) 
           OR %s NOT IN (SELECT id FROM %s)
    `, tableName, leftCol, leftTable, rightCol, rightTable)

	result := tx.Exec(query)
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup %s: %w", tableName, result.Error)
	}

	if result.RowsAffected > 0 {
		log.Warn().
			Str("table", tableName).
			Int64("rows_deleted", result.RowsAffected).
			Msg("cleaned up orphaned many-to-many relations")
	}

	return nil
}
