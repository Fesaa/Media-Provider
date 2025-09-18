package manual

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func AssignPreferencesToFirstAdmin(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	firstAdmin, err := getDefaultUser(ctx, db)
	if err != nil {
		return err
	}

	var preferences models.UserPreferences
	if err = db.WithContext(ctx).First(&preferences).Error; err != nil {
		return allowNoRecord(err) // Tests may not have preferences loaded, default preferences will be assigned later
	}

	preferences.UserID = firstAdmin.ID
	return db.WithContext(ctx).Save(&preferences).Error
}

func SetDefaultUserPreferences(ctx context.Context, db *gorm.DB, log zerolog.Logger) error {
	var userPreferences []models.UserPreferences
	var users []models.User

	if err := db.WithContext(ctx).Find(&userPreferences).Error; err != nil {
		return err
	}

	if err := db.WithContext(ctx).Find(&users).Error; err != nil {
		return err
	}

	userPreferencesDict := make(map[int]*models.UserPreferences)
	for _, userPref := range userPreferences {
		userPreferencesDict[userPref.UserID] = &userPref
	}

	toAdd := make([]models.UserPreferences, 0)

	for _, user := range users {
		_, ok := userPreferencesDict[user.ID]
		if ok {
			continue
		}

		toAdd = append(toAdd, models.UserPreferences{
			UserID:              user.ID,
			LogEmptyDownloads:   false,
			ConvertToWebp:       true,
			CoverFallbackMethod: models.CoverFallbackLast,
			GenreList:           []string{},
			BlackList:           []string{},
			WhiteList:           []string{},
			AgeRatingMappings:   []models.AgeRatingMapping{},
			TagMappings:         []models.TagMapping{},
		})
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, userPref := range toAdd {
			if err := tx.Create(&userPref).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
