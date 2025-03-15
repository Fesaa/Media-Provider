package impl

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

type preferences struct {
	db *gorm.DB
}

func Preferences(db *gorm.DB) models.Preferences {
	return &preferences{db: db}
}

func (p preferences) Get() (*models.Preference, error) {
	var pref models.Preference
	err := p.db.First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (p preferences) GetComplete() (*models.Preference, error) {
	var pref models.Preference
	err := p.db.
		Preload("DynastyGenreTags").
		Preload("BlackListedTags").
		Preload("AgeRatingMappings").
		Preload("AgeRatingMappings.Tag").
		First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (p preferences) Update(pref models.Preference) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&pref).Error; err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("DynastyGenreTags").Replace(pref.DynastyGenreTags); err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("BlackListedTags").Replace(pref.BlackListedTags); err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("AgeRatingMappings").Replace(pref.AgeRatingMappings); err != nil {
			return err
		}

		return nil
	})
}
