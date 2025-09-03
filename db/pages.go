package db

import (
	"errors"

	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

type pages struct {
	db *gorm.DB
}

func Pages(db *gorm.DB) models.Pages {
	return &pages{db}
}

func (p *pages) All() ([]models.Page, error) {
	var allPages []models.Page
	result := p.db.Preload("Modifiers", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Preload("Modifiers.Values").Find(&allPages)
	if result.Error != nil {
		return nil, result.Error
	}

	return allPages, nil
}

func (p *pages) Get(id uint) (*models.Page, error) {
	var page models.Page
	result := p.db.Preload("Modifiers", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Preload("Modifiers.Values").First(&page, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}
	return &page, nil
}

func (p *pages) New(page *models.Page) error {
	return p.db.Create(page).Error
}

func (p *pages) Update(page *models.Page) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(page).Error; err != nil {
			return err
		}

		if err := tx.Model(&page).Association("Modifiers").Replace(page.Modifiers); err != nil {
			return err
		}

		for _, modifier := range page.Modifiers {
			if err := tx.Model(&modifier).Association("Values").Replace(modifier.Values); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *pages) UpdateMany(pages []models.Page) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		for _, page := range pages {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(page).Error; err != nil {
				return err
			}

			if err := tx.Model(&page).Association("Modifiers").Replace(page.Modifiers); err != nil {
				return err
			}

			for _, modifier := range page.Modifiers {
				if err := tx.Model(&modifier).Association("Values").Replace(modifier.Values); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (p *pages) Delete(id uint) error {
	return p.db.Delete(&models.Page{}, id).Error
}
