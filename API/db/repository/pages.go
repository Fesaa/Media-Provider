package repository

import (
	"context"
	"errors"

	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/devfeel/mapper"
	"gorm.io/gorm"
)

type PagesRepository interface {
	GetAllPages(context.Context) ([]models.Page, error)
	GetPage(context.Context, int) (*models.Page, error)

	Create(context.Context, *models.Page) error
	Update(context.Context, *models.Page) error
	UpdateMany(context.Context, []models.Page) error
	Delete(context.Context, int) error
}

type pagesRepository struct {
	db     *gorm.DB
	mapper mapper.IMapper
}

func (p pagesRepository) GetAllPages(ctx context.Context) ([]models.Page, error) {
	var pages []models.Page
	res := p.db.WithContext(ctx).Preload("Modifiers", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Preload("Modifiers.Values").Find(&pages)
	if res.Error != nil {
		return nil, res.Error
	}

	return pages, nil
}

func (p pagesRepository) GetPage(ctx context.Context, id int) (*models.Page, error) {
	var page models.Page
	result := p.db.WithContext(ctx).Preload("Modifiers", func(db *gorm.DB) *gorm.DB {
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

func (p pagesRepository) Create(ctx context.Context, page *models.Page) error {
	page.ID = 0
	return p.db.WithContext(ctx).Create(page).Error
}

func (p pagesRepository) Update(ctx context.Context, page *models.Page) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (p pagesRepository) UpdateMany(ctx context.Context, pages []models.Page) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (p pagesRepository) Delete(ctx context.Context, id int) error {
	return p.db.WithContext(ctx).Delete(&models.Page{}, id).Error
}

func NewPagesRepository(db *gorm.DB, m mapper.IMapper) PagesRepository {
	return &pagesRepository{db: db, mapper: m}
}
