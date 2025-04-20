package models

import "time"

type Pages interface {
	All() ([]Page, error)
	Get(id uint) (*Page, error)

	New(page *Page) error
	Update(page *Page) error

	Delete(id uint) error
}

type Notifications interface {
	Get(id uint) (Notification, error)
	GetMany(ids []uint) ([]Notification, error)
	All() ([]Notification, error)
	AllAfter(time.Time) ([]Notification, error)

	New(Notification) error
	Delete(uint) error
	DeleteMany([]uint) error

	MarkRead(uint) error
	MarkReadMany([]uint) error
	MarkUnread(uint) error
	Unread() (int64, error)
}

type Preferences interface {
	// Get returns a pointer to Preference, with no relations loaded
	Get() (*Preference, error)
	// GetComplete returns a pointer to Preference, with all relations loaded
	GetComplete() (*Preference, error)
	Update(pref Preference) error
	// Flush set the cached value for Get and GetComplete to nil
	Flush() error
}

type Subscriptions interface {
	All() ([]Subscription, error)
	Get(uint) (*Subscription, error)
	GetByContentId(string) (*Subscription, error)

	New(Subscription) (*Subscription, error)
	Update(Subscription) error
	Delete(uint) error
}

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
	GetResetByUserId(userId uint) (*PasswordReset, error)
	GetReset(key string) (*PasswordReset, error)
	DeleteReset(key string) error

	Delete(id uint) error
}

type Metadata interface {
	All() ([]MetadataRow, error)
	GetRow(key MetadataKey) (*MetadataRow, error)
	UpdateRow(metadata MetadataRow) error
	Update([]MetadataRow) error
}

type Option[T any] func(T) T
