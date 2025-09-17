package manual

import (
	"context"
	"strings"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func AssignMissingRelations(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	type ageRatingMapRelation struct {
		PreferenceID   int
		AgeRatingMapID int
	}

	type tag struct {
		ID             int
		AgeRatingMapID int
	}

	var all []ageRatingMapRelation
	if err := db.WithContext(ctx).Table("preference_age_rating_mappings").Find(&all).Error; err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return nil
		}
		return err
	}

	var allTags []tag
	if err := db.WithContext(ctx).Table("tags").Find(&allTags).Error; err != nil {
		if strings.Contains(err.Error(), "no such table") {
			return nil
		}
		return err
	}

	var allMappings []models.AgeRatingMap
	if err := db.WithContext(ctx).Find(&allMappings).Error; err != nil {
		return err
	}

	findRel := func(mapping models.AgeRatingMap) (ageRatingMapRelation, bool) {
		for _, rel := range all {
			if rel.AgeRatingMapID == mapping.ID {
				return rel, true
			}
		}

		return ageRatingMapRelation{}, false
	}

	findTagRel := func(mapping models.AgeRatingMap) (tag, bool) {
		for _, rel := range allTags {
			if rel.AgeRatingMapID == mapping.ID {
				return rel, true
			}
		}
		return tag{}, false
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, mapping := range allMappings {
			rel, ok := findRel(mapping)
			if !ok {
				log.Warn().Int("id", mapping.ID).Msg("mapping not mapped, removing")
				if err := tx.Delete(&mapping).Error; err != nil {
					return err
				}
				continue
			}

			mapping.PreferenceID = rel.PreferenceID

			tagRel, ok := findTagRel(mapping)
			if ok {
				mapping.TagID = tagRel.ID
			}

			if err := tx.Save(&mapping).Error; err != nil {
				return err
			}
		}

		return nil
	})

}
