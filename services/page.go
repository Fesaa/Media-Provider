package services

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var (
	ErrPageNotFound       = errors.New("page not found")
	ErrExistingPagesFound = errors.New("some pages already exists")
	ErrFailedToSortCheck  = errors.New("error during sort checks")

	DefaultPageSort = 9999
)

type PageService interface {
	UpdateOrCreate(page *models.Page) error
	SwapPages(int64, int64) error
	LoadDefaultPages() error
}

type pageService struct {
	db  *db.Database
	log zerolog.Logger
}

func PageServiceProvider(db *db.Database, log zerolog.Logger) PageService {
	return &pageService{
		db:  db,
		log: log.With().Str("handler", "page-service").Logger(),
	}
}

func (ps *pageService) UpdateOrCreate(page *models.Page) error {
	var other models.Page
	err := ps.db.DB().
		Not(models.Page{Model: gorm.Model{ID: page.ID}}).
		First(&other, models.Page{SortValue: page.SortValue}).
		Error
	// Must return gorm.ErrRecordNotFound
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		ps.log.Error().Err(err).Msg("Error occurred during sort check")
		return ErrFailedToSortCheck
	}

	if page.SortValue == DefaultPageSort {
		var maxPageSort int
		err = ps.db.DB().Model(&models.Page{}).Select("MAX(sort_value) AS maxPageSort").Scan(&maxPageSort).Error
		if err != nil {
			ps.log.Error().Err(err).Msg("Error occurred while getting max page sort")
			return ErrFailedToSortCheck
		}
		page.SortValue = maxPageSort + 1
	}

	return ps.db.Pages.Update(page)
}

func (ps *pageService) SwapPages(id1, id2 int64) error {
	page1, err := ps.db.Pages.Get(id1)
	if err != nil {
		ps.log.Error().Err(err).Int64("id", id1).Msg("Failed to get page1")
		return ErrPageNotFound
	}
	page2, err := ps.db.Pages.Get(id2)
	if err != nil {
		ps.log.Error().Err(err).Int64("id", id2).Msg("Failed to get page2")
		return ErrPageNotFound
	}

	page1.SortValue, page2.SortValue = page2.SortValue, page1.SortValue

	err = ps.db.DB().Transaction(func(tx *gorm.DB) error {
		if err = tx.Save(page1).Error; err != nil {
			return err
		}

		if err = tx.Save(page2).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		ps.log.Error().Err(err).
			Int64("id1", id1).
			Int64("id2", id2).
			Msg("Failed to swap pages")
		return fmt.Errorf("failed to swap pages: %w", err)
	}

	return nil
}

func (ps *pageService) LoadDefaultPages() error {
	pages, err := ps.db.Pages.All()
	if err != nil {
		ps.log.Error().Err(err).Msg("Failed to load existing pages, not loading default pages")
		return err
	}

	if len(pages) != 0 {
		return ErrExistingPagesFound
	}

	return ps.db.DB().Transaction(func(tx *gorm.DB) error {
		for _, page := range models.DefaultPages {
			if err = tx.Create(&page).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
