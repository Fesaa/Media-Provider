package repository

import (
	"context"
	"time"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/utils"
	"gorm.io/gorm"
)

type PreferencesRepository interface {
	// GetPreferences returns the models.UserPreferences for the given user
	GetPreferences(context.Context, int) (*models.UserPreferences, error)
	// Update saves the given models.UserPreferences
	Update(context.Context, *models.UserPreferences) error
}

type preferencesRepository struct {
	db    *gorm.DB
	cache utils.SafeMap[int, utils.CachedItem[*models.UserPreferences]]
}

func getFromCache(cache utils.SafeMap[int, utils.CachedItem[*models.UserPreferences]], key int) *models.UserPreferences {
	cachedItem, ok := cache.Get(key)
	if !ok {
		return nil
	}

	if cachedItem.HasExpired() {
		cache.Delete(key)
		return nil
	}

	item, err := cachedItem.Get()
	if err != nil {
		return nil
	}

	return item
}

func (p *preferencesRepository) GetPreferences(ctx context.Context, userId int) (*models.UserPreferences, error) {
	if cache := getFromCache(p.cache, userId); cache != nil {
		return cache, nil
	}

	var pref models.UserPreferences
	err := p.db.WithContext(ctx).First(&pref, &models.UserPreferences{UserID: userId}).Error
	if err != nil {
		return nil, err
	}

	p.cache.Set(userId, utils.NewCachedItem(&pref, 5*time.Minute))
	return &pref, nil
}

func (p *preferencesRepository) Update(ctx context.Context, pref *models.UserPreferences) error {
	if err := p.db.WithContext(ctx).Save(pref).Error; err != nil {
		return err
	}

	p.cache.Delete(pref.UserID)
	return nil
}

func NewPreferencesRepository(db *gorm.DB) PreferencesRepository {
	return &preferencesRepository{
		db:    db,
		cache: utils.NewSafeMap[int, utils.CachedItem[*models.UserPreferences]](),
	}
}
