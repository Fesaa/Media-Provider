package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"slices"
)

var (
	ErrPageNotFound       = errors.New("page not found")
	ErrExistingPagesFound = errors.New("some pages already exists")
	ErrFailedToSortCheck  = errors.New("error during sort checks")

	DefaultPageSort = 9999
)

type PageService interface {
	UpdateOrCreate(page *models.Page) error
	OrderPages([]uint) error
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
		Not(models.Page{Model: models.Model{ID: page.ID}}).
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

func (ps *pageService) OrderPages(order []uint) error {
	pages, err := ps.db.Pages.All()
	if err != nil {
		return err
	}

	newPages := make([]models.Page, len(pages))

	for sliceIdx, page := range pages {
		idx := slices.Index(order, page.ID)
		if idx == -1 {
			return ErrFailedToSortCheck
		}

		page.SortValue = idx

		newPages[sliceIdx] = page
	}

	return ps.db.Pages.UpdateMany(newPages)
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
