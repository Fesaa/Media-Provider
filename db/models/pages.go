package models

type Pages interface {
	All() ([]Page, error)
	Get(id int64) (*Page, error)

	New(page Page) error
	Update(page Page) error

	Delete(id int64) error
}
