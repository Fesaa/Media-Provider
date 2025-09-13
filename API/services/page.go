package services

import (
	"context"
	"database/sql"
	"errors"
	"slices"

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
	UpdateOrCreate(context.Context, *models.Page) error
	OrderPages(context.Context, []int) error
	LoadDefaultPages(context.Context) error
}

type pageService struct {
	unitOfWork *db.UnitOfWork
	log        zerolog.Logger
}

func PageServiceProvider(unitOfWork *db.UnitOfWork, log zerolog.Logger) PageService {
	return &pageService{
		unitOfWork: unitOfWork,
		log:        log.With().Str("handler", "page-service").Logger(),
	}
}

func (ps *pageService) UpdateOrCreate(ctx context.Context, page *models.Page) error {
	var other models.Page

	err := ps.unitOfWork.DB().
		WithContext(ctx).
		Not(models.Page{Model: models.Model{ID: page.ID}}).
		Where(map[string]interface{}{"sort_value": page.SortValue}). // https://gorm.io/docs/query.html#Struct-amp-Map-Conditions
		First(&other).
		Error
	// Must return gorm.ErrRecordNotFound
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		if page.ID > 0 {
			// While this should never happen, forcefully reset the sort is better than having the page be un-editable.
			ps.log.Error().Str("other", other.Title).Err(err).
				Msg("Unwanted error, or found matching sort value, resetting to default value")
		}

		page.SortValue = DefaultPageSort
	}

	if page.SortValue == DefaultPageSort {
		var maxPageSort sql.NullInt64
		err = ps.unitOfWork.DB().
			WithContext(ctx).
			Model(&models.Page{}).
			Select("MAX(sort_value) AS maxPageSort").
			Scan(&maxPageSort).Error
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

	if page.ID < 0 {
		return ps.unitOfWork.Pages.Create(ctx, page)
	}

	return ps.unitOfWork.Pages.Update(ctx, page)
}

func (ps *pageService) OrderPages(ctx context.Context, order []int) error {
	pages, err := ps.unitOfWork.Pages.GetAllPages(ctx)
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

	return ps.unitOfWork.Pages.UpdateMany(ctx, newPages)
}

func (ps *pageService) LoadDefaultPages(ctx context.Context) error {
	pages, err := ps.unitOfWork.Pages.GetAllPages(ctx)
	if err != nil {
		ps.log.Error().Err(err).Msg("Failed to load existing pages, not loading default pages")
		return err
	}

	if len(pages) != 0 {
		return ErrExistingPagesFound
	}

	return ps.unitOfWork.DB().Transaction(func(tx *gorm.DB) error {
		for _, page := range models.DefaultPages {
			if err = tx.Create(&page).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
