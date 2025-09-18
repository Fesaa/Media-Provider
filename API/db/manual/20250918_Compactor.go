package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type modifier struct {
	models.Model

	PageID int

	Title  string              `json:"title"`
	Type   models.ModifierType `json:"type"`
	Key    string              `json:"key"`
	Values []modifierValue     `json:"values"`
	Sort   int
}

type modifierValue struct {
	models.Model

	ModifierID int
	Key        string `json:"key"`
	Value      string `json:"value"`
	Default    bool   `json:"default"`
}

// PageCompactor will compact all modifiers and modifier values into a json column for pages
func PageCompactor(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	var pages []models.Page
	var modifiers []modifier
	var modifierValues []modifierValue

	if err := db.WithContext(ctx).Find(&pages).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Find(&modifiers).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Find(&modifierValues).Error; err != nil {
		return allowNoTable(err)
	}

	modifierDict := make(map[int]*modifier)
	for _, m := range modifiers {
		modifierDict[m.ID] = &m
	}

	for _, mv := range modifierValues {
		m, ok := modifierDict[mv.ModifierID]
		if !ok {
			log.Warn().Int("id", mv.ID).Msg("Modifier value without modifier found, won't be migrated")
			continue
		}

		m.Values = append(m.Values, mv)
	}

	pageDict := make(map[int]*models.Page)
	for _, page := range pages {
		pageDict[page.ID] = &page
	}

	for _, m := range modifierDict {
		p, ok := pageDict[m.ID]
		if !ok {
			log.Warn().Int("id", m.ID).Msg("Modifier without page found, won't be migrated")
			continue
		}

		p.Modifiers = append(p.Modifiers, models.Modifier{
			Title: m.Title,
			Type:  m.Type,
			Key:   m.Key,
			Values: utils.Map(m.Values, func(mv modifierValue) models.ModifierValue {
				return models.ModifierValue{
					Key:     mv.Key,
					Value:   mv.Value,
					Default: mv.Default,
				}
			}),
			Sort: m.Sort,
		})
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, page := range pageDict {
			if err := tx.Save(page).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err = db.Migrator().DropTable("modifiers"); err != nil {
			return err
		}

		return db.Migrator().DropTable("modifier_values")
	})
}
