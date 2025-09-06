package models

import "time"

type Pages interface {
	All() ([]Page, error)
	Get(id int) (*Page, error)

	New(page *Page) error
	Update(page *Page) error
	UpdateMany(pages []Page) error

	Delete(id int) error
}

type Notifications interface {
	Get(id int) (Notification, error)
	GetMany(ids []int) ([]Notification, error)
	All() ([]Notification, error)
	AllAfter(time.Time) ([]Notification, error)
	Recent(int, NotificationGroup) ([]Notification, error)

	New(Notification) error
	Delete(int) error
	DeleteMany([]int) error

	MarkRead(int) error
	MarkReadMany([]int) error
	MarkUnread(int) error
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
	AllForUser(int) ([]Subscription, error)
	Get(int) (*Subscription, error)
	GetForUser(int, int) (Subscription, error)
	GetByContentId(string) (*Subscription, error)
	GetByContentIdForUser(string, int) (*Subscription, error)

	New(Subscription) (*Subscription, error)
	Update(Subscription) error
	Delete(int) error
}

type Users interface {
	All() ([]User, error)
	ExistsAny() (bool, error)

	GetById(id int) (*User, error)
	GetByExternalId(ExternalId string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByName(name string) (*User, error)
	GetByApiKey(key string) (*User, error)

	Create(name string, opts ...Option[User]) (*User, error)
	Update(user User, opts ...Option[User]) (*User, error)
	UpdateById(id int, opts ...Option[User]) (*User, error)

	GenerateReset(userId int) (*PasswordReset, error)
	GetResetByUserId(userId int) (*PasswordReset, error)
	GetReset(key string) (*PasswordReset, error)
	DeleteReset(key string) error

	Delete(id int) error
}

type Metadata interface {
	All() ([]MetadataRow, error)
	GetRow(key MetadataKey) (*MetadataRow, error)
	UpdateRow(metadata MetadataRow) error
	Update([]MetadataRow) error
}

type Settings interface {
	All() ([]ServerSetting, error)
	GetById(SettingKey) (ServerSetting, error)
	Update([]ServerSetting) error
}

type Option[T any] func(T) T
