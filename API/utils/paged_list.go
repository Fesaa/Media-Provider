package utils

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	DefaultPageSize = 10
)

type UserParams struct {
	PageSize   int `query:"pageSize"`
	PageNumber int `query:"pageNumber"`
}

type PagedList[T any] struct {
	Items       []T `json:"items"`
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

func NewPageListFromUserParams[T any](ctx context.Context, query *gorm.DB, params UserParams) (PagedList[T], error) {
	return NewPagedList[T](ctx, query, params.PageNumber, params.PageSize)
}

func NewPagedList[T any](ctx context.Context, query *gorm.DB, page, pageSize int) (PagedList[T], error) {
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}

	var count int64
	if err := query.WithContext(ctx).Count(&count).Error; err != nil {
		return PagedList[T]{}, fmt.Errorf("paged list count: %w", err)
	}

	if _, ok := query.Statement.Clauses[clause.OrderBy{}.Name()]; !ok {
		var empty T
		log.Warn().Str("handler", "NewPagedList").
			Msgf("Paged list (%T) created without an ORDER BY clause, results may be inconsistent.", empty)
	}

	pg := PagedList[T]{
		CurrentPage: page,
		PageSize:    pageSize,
	}

	offSet := page * pageSize
	if offSet > int(count) { // Fewer items in list than available
		return pg, nil
	}

	var items []T
	if err := query.WithContext(ctx).Offset(offSet).Limit(pageSize).Find(&items).Error; err != nil {
		return PagedList[T]{}, fmt.Errorf("paged list items: %w", err)
	}

	pg.Items = items
	return pg, nil
}
