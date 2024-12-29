package models

type Users interface {
	All() ([]User, error)
	ExistsAny() (bool, error)

	GetById(id uint) (*User, error)
	GetByName(name string) (*User, error)
	GetByApiKey(key string) (*User, error)

	Create(name string, opts ...Option[User]) (*User, error)
	Update(user User, opts ...Option[User]) (*User, error)
	UpdateById(id uint, opts ...Option[User]) (*User, error)

	GenerateReset(userId uint) (*PasswordReset, error)
	GetReset(key string) (*PasswordReset, error)
	DeleteReset(key string) error

	Delete(id uint) error
}

type Option[T any] func(T) T
