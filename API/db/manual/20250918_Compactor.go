package manual

import (
	"context"
	"fmt"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/comicinfo"
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

type oldPreference struct {
	models.Model

	SubscriptionRefreshHour int                        `json:"subscriptionRefreshHour" validate:"min=0,max=23"`
	LogEmptyDownloads       bool                       `json:"logEmptyDownloads" validate:"boolean"`
	ConvertToWebp           bool                       `json:"convertToWebp" validate:"boolean"`
	CoverFallbackMethod     models.CoverFallbackMethod `json:"coverFallbackMethod"`
}

type tag struct {
	models.Model

	PreferenceID   int
	AgeRatingMapID int

	Name           string `json:"name"`
	NormalizedName string `json:"normalizedName"`
}

type ageRatingMap struct {
	models.Model

	PreferenceID       int
	Tag                tag                 `json:"tag"`
	ComicInfoAgeRating comicinfo.AgeRating `json:"comicInfoAgeRating"`
	// MetronAgeRating    metroninfo.AgeRating `json:"metronAgeRating"`
}

type tagMap struct {
	models.Model

	PreferenceID int
	OriginID     int
	Origin       tag `json:"origin" gorm:"foreignKey:OriginID;references:ID"`
	DestID       int
	Dest         tag `json:"dest" gorm:"foreignKey:DestID;references:ID"`
}

type mapping struct {
	PreferenceID int
	TagID        int
}

type ageRatingMapping struct {
	PreferenceID   int
	AgeRatingMapID int
}

func PreferenceCompactor(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	var preference oldPreference
	var tags []tag
	var ageRatingMaps []ageRatingMap
	var tagMaps []tagMap

	var genreMappings []mapping
	var blackListMappings []mapping
	var whiteListMappings []mapping
	var ageRatingMappings []ageRatingMapping

	if err := db.WithContext(ctx).Table("preferences").First(&preference).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Find(&tags).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Find(&ageRatingMaps).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Find(&tagMaps).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Table("preference_dynasty_genre_tags").Find(&genreMappings).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Table("preference_black_list_tags").Find(&blackListMappings).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Table("preference_white_list_tags").Find(&whiteListMappings).Error; err != nil {
		return allowNoTable(err)
	}

	if err := db.WithContext(ctx).Table("preference_age_rating_mappings").Find(&ageRatingMappings).Error; err != nil {
		return allowNoTable(err)
	}

	tagsDict := make(map[int]tag)
	for _, t := range tags {
		tagsDict[t.ID] = t
	}

	ageRatingMapsDict := make(map[int]ageRatingMap)
	for _, t := range ageRatingMaps {
		ageRatingMapsDict[t.ID] = ageRatingMap{}
	}

	filter := func(mappings []mapping) []tag {
		out := make([]tag, 0)
		for _, m := range mappings {
			t, ok := tagsDict[m.TagID]
			if ok {
				out = append(out, t)
			}
		}
		return out
	}

	preferences := models.UserPreferences{
		LogEmptyDownloads:   preference.LogEmptyDownloads,
		ConvertToWebp:       preference.ConvertToWebp,
		CoverFallbackMethod: preference.CoverFallbackMethod,
		GenreList: utils.Map(filter(genreMappings), func(t tag) string {
			return t.Name
		}),
		BlackList: utils.Map(filter(blackListMappings), func(t tag) string {
			return t.Name
		}),
		WhiteList: utils.Map(filter(whiteListMappings), func(t tag) string {
			return t.Name
		}),
		AgeRatingMappings: utils.MaybeMap(ageRatingMappings, func(t ageRatingMapping) (models.AgeRatingMapping, bool) {
			arm, ok := ageRatingMapsDict[t.AgeRatingMapID]
			if !ok {
				return models.AgeRatingMapping{}, false
			}

			armTag, ok := utils.FindOk(tags, func(t tag) bool {
				return t.AgeRatingMapID == arm.ID
			})
			if !ok {
				return models.AgeRatingMapping{}, false
			}

			return models.AgeRatingMapping{
				Tag:                armTag.Name,
				ComicInfoAgeRating: arm.ComicInfoAgeRating,
			}, true
		}),
		TagMappings: utils.MaybeMap(tagMaps, func(tm tagMap) (models.TagMapping, bool) {
			origin, ok := tagsDict[tm.OriginID]
			if !ok {
				return models.TagMapping{}, false
			}

			dest, ok := tagsDict[tm.DestID]
			if !ok {
				return models.TagMapping{}, false
			}

			return models.TagMapping{
				OriginTag:      origin.Name,
				DestinationTag: dest.Name,
			}, true
		}),
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&preferences).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.ServerSetting{
			Key:   models.SubscriptionRefreshHour,
			Value: fmt.Sprintf("%d", preference.SubscriptionRefreshHour),
		}).Error; err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("age_rating_maps"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("preference_dynasty_genre_tags"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("preference_white_list_tags"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("preference_age_rating_mappings"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("preference_dynasty_genre_tags"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("preferences"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("tags"); err != nil {
			return err
		}

		if err := tx.Migrator().DropTable("tag_maps"); err != nil {
			return err
		}

		return nil
	})
}
