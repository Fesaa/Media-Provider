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
	res := p.db.WithContext(ctx).Find(&pages)
	if res.Error != nil {
		return nil, res.Error
	}

	return pages, nil
}

func (p pagesRepository) GetPage(ctx context.Context, id int) (*models.Page, error) {
	var page models.Page
	result := p.db.WithContext(ctx).First(&page, id)
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
	return p.db.WithContext(ctx).Save(page).Error
}

func (p pagesRepository) UpdateMany(ctx context.Context, pages []models.Page) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, page := range pages {
			if err := tx.Save(&page).Error; err != nil {
				return err
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
