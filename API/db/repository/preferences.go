package repository

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/devfeel/mapper"
	"gorm.io/gorm"
)

type PreferencesRepository interface {
	// GetPreferences returns a pointer to Preference, with no relations loaded
	GetPreferences(context.Context) (*models.Preference, error)
	// GetPreferencesComplete returns a pointer to Preference, with all relations loaded
	GetPreferencesComplete(context.Context) (*models.Preference, error)
	Update(context.Context, *models.Preference) error
	// Flush flushes the cached values of GetPreferences and GetPreferencesComplete
	Flush() error
}

type preferencesRepository struct {
	db             *gorm.DB
	mapper         mapper.IMapper
	cachedComplete utils.CachedItem[*models.Preference]
	cachedSlim     utils.CachedItem[*models.Preference]
}

func (p *preferencesRepository) GetPreferences(ctx context.Context) (*models.Preference, error) {
	if p.cachedSlim != nil && !p.cachedSlim.HasExpired() {
		return p.cachedSlim.Get()
	}

	var pref models.Preference
	err := p.db.WithContext(ctx).First(&pref).Error
	if err != nil {
		return nil, err
	}

	p.cachedSlim = utils.NewCachedItem(&pref, 5*time.Minute)
	return &pref, nil
}

func (p *preferencesRepository) GetPreferencesComplete(ctx context.Context) (*models.Preference, error) {
	if p.cachedComplete != nil && !p.cachedComplete.HasExpired() {
		return p.cachedComplete.Get()
	}

	var pref models.Preference
	err := p.db.
		WithContext(ctx).
		Preload("DynastyGenreTags").
		Preload("BlackListedTags").
		Preload("WhiteListedTags").
		Preload("AgeRatingMappings").
		Preload("AgeRatingMappings.Tag").
		Preload("TagMappings").
		Preload("TagMappings.Origin").
		Preload("TagMappings.Dest").
		First(&pref).Error
	if err != nil {
		return nil, err
	}

	p.cachedComplete = utils.NewCachedItem(&pref, 5*time.Minute)
	return &pref, nil
}

func (p *preferencesRepository) Update(ctx context.Context, pref *models.Preference) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		if err := tx.Model(&pref).Association("TagMappings").Replace(pref.TagMappings); err != nil {
			return err
		}

		return nil
	})
}

func (p *preferencesRepository) Flush() error {
	p.cachedComplete = nil
	p.cachedSlim = nil
	return nil
}

func NewPreferencesRepository(db *gorm.DB, m mapper.IMapper) PreferencesRepository {
	return &preferencesRepository{db: db, mapper: m}
}
