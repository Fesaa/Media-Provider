package impl

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"gorm.io/gorm"
)

type pagesImpl struct {
	db *gorm.DB
}

func Pages(db *gorm.DB) models.Pages {
	return &pagesImpl{db}
}

func (p *pagesImpl) All() ([]models.Page, error) {
	var pages []models.Page
	result := p.db.Preload("Modifiers").Preload("Modifiers.Values").Find(&pages)
	if result.Error != nil {
		return nil, result.Error
	}

	return pages, nil
}

func (p *pagesImpl) Get(id int64) (*models.Page, error) {
	var page models.Page
	result := p.db.Preload("Modifiers").Preload("Modifiers.Values").First(&page, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &page, nil
}

func (p *pagesImpl) New(page models.Page) error {
	return p.db.Create(&page).Error
}

func (p *pagesImpl) Update(page models.Page) error {
	return p.db.Save(&page).Error
}

func (p *pagesImpl) Delete(id int64) error {
	return p.db.Delete(&models.Page{}, id).Error
}
