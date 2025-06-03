package impl

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
	"time"
)

type preferences struct {
	db             *gorm.DB
	cachedComplete utils.CachedItem[*models.Preference]
	cachedSlim     utils.CachedItem[*models.Preference]
}

func Preferences(db *gorm.DB) models.Preferences {
	return &preferences{db: db}
}

func (p *preferences) Flush() error {
	p.cachedComplete = nil
	p.cachedSlim = nil
	return nil
}

func (p *preferences) Get() (*models.Preference, error) {
	if p.cachedSlim != nil && !p.cachedSlim.HasExpired() {
		return p.cachedSlim.Get()
	}

	var pref models.Preference
	err := p.db.First(&pref).Error
	if err != nil {
		return nil, err
	}

	p.cachedSlim = utils.NewCachedItem(&pref, 5*time.Minute)
	return &pref, nil
}

func (p *preferences) GetComplete() (*models.Preference, error) {
	if p.cachedComplete != nil && !p.cachedComplete.HasExpired() {
		return p.cachedComplete.Get()
	}

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

	p.cachedComplete = utils.NewCachedItem(&pref, 5*time.Minute)
	return &pref, nil
}

func (p *preferences) Update(pref models.Preference) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Select("*").Updates(&pref).Error; err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("DynastyGenreTags").Replace(pref.DynastyGenreTags); err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("BlackListedTags").Replace(pref.BlackListedTags); err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("WhiteListedTags").Replace(pref.WhiteListedTags); err != nil {
			return err
		}

		if err := tx.Model(&pref).Association("AgeRatingMappings").Replace(pref.AgeRatingMappings); err != nil {
			return err
		}

		return nil
	})
}
