package api

import "github.com/Fesaa/Media-Provider/db/models"

type Users interface {
	All() ([]models.User, error)
	ExistsAny() (bool, error)

	GetById(id int64) (*models.User, error)
	GetByName(name string) (*models.User, error)
	GetByApiKey(key string) (*models.User, error)

	Create(name string, opts ...models.Option[models.User]) (*models.User, error)
	Update(user models.User, opts ...models.Option[models.User]) (*models.User, error)
	UpdateById(id int64, opts ...models.Option[models.User]) (*models.User, error)

	GenerateReset(userId int64) (*models.PassWordReset, error)
	GetReset(key string) (*models.PassWordReset, error)
	DeleteReset(key string) error

	Delete(id int64) error
}
