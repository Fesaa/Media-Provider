package repository

import (
	"context"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/devfeel/mapper"
	"gorm.io/gorm"
)

type SettingsRepository interface {
	GetAll(context.Context) ([]models.ServerSetting, error)
	Update(context.Context, []models.ServerSetting) error
}

type settingsRepository struct {
	db     *gorm.DB
	mapper mapper.IMapper
}

func (s settingsRepository) GetAll(ctx context.Context) ([]models.ServerSetting, error) {
	var rows []models.ServerSetting
	if err := s.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s settingsRepository) Update(ctx context.Context, settings []models.ServerSetting) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, setting := range settings {
			if err := tx.Where("key = ?", setting.Key).Save(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func NewSettingsRepository(db *gorm.DB, m mapper.IMapper) SettingsRepository {
	return &settingsRepository{db: db, mapper: m}
}
