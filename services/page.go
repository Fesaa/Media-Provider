package services

import (
	"database/sql"
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
	SwapPages(uint, uint) error
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
		Where(map[string]interface{}{"sort_value": 0}). // https://gorm.io/docs/query.html#Struct-amp-Map-Conditions
		First(&other).
		Error
	// Must return gorm.ErrRecordNotFound
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		ps.log.Error().Str("other", fmt.Sprintf("%+v", other)).Err(err).
			Msg("Unwanted error, or found matching sort value, resetting to default value")
		// While this should never happen, forcefully reset the sort is better than having the page be un-editable.
		page.SortValue = DefaultPageSort
	}

	if page.SortValue == DefaultPageSort {
		var maxPageSort sql.NullInt64
		err = ps.db.DB().Model(&models.Page{}).Select("MAX(sort_value) AS maxPageSort").Scan(&maxPageSort).Error
		if err != nil {
			ps.log.Error().Err(err).Msg("Error occurred while getting max page sort")
			return ErrFailedToSortCheck
		}

		if maxPageSort.Valid {
			page.SortValue = int(maxPageSort.Int64) + 1
		} else {
			page.SortValue = 0 // First page being inserted
		}
	}

	return ps.db.Pages.Update(page)
}

func (ps *pageService) SwapPages(id1, id2 uint) error {
	page1, err := ps.db.Pages.Get(id1)
	if err != nil {
		ps.log.Error().Err(err).Uint("id", id1).Msg("Failed to get page1")
		return ErrPageNotFound
	}
	page2, err := ps.db.Pages.Get(id2)
	if err != nil {
		ps.log.Error().Err(err).Uint("id", id2).Msg("Failed to get page2")
		return ErrPageNotFound
	}

	if page1 == nil || page2 == nil {
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
			Uint("id1", id1).
			Uint("id2", id2).
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
