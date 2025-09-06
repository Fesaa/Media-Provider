package db

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

type settings struct {
	db *gorm.DB
}

func Settings(db *gorm.DB) models.Settings {
	return &settings{db}
}

func (s *settings) All() ([]models.ServerSetting, error) {
	var rows []models.ServerSetting
	if err := s.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *settings) GetById(id models.SettingKey) (models.ServerSetting, error) {
	var row models.ServerSetting
	if err := s.db.First(&row, id).Error; err != nil {
		var zero models.ServerSetting
		return zero, err
	}
	return row, nil
}

func (s *settings) Update(serverSettings []models.ServerSetting) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, setting := range serverSettings {
			if err := tx.Where("key = ?", setting.Key).Save(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
