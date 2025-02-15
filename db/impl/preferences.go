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

func (p preferences) GetWithTags() (*models.Preference, error) {
	var pref models.Preference
	err := p.db.
		Preload("DynastyGenreTags").
		Preload("BlackListedTags").
		First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (p preferences) Update(pref models.Preference) error {
	return p.db.Save(&pref).Error
}
