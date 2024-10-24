package api

import "github.com/Fesaa/Media-Provider/db/models"

type Pages interface {
	All() ([]models.Page, error)
	Get(id int64) (*models.Page, error)

	Upsert(pages ...*models.Page) error

	Delete(id int64) error
}
