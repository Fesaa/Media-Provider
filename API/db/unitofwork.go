package db

import (
	"database/sql"

	"github.com/Fesaa/Media-Provider/db/repository"
	"gorm.io/gorm"
)

type UnitOfWork struct {
	db *gorm.DB

	isTx bool

	Pages         repository.PagesRepository
	Subscriptions repository.SubscriptionsRepository
	Preferences   repository.PreferencesRepository
	Notifications repository.NotificationsRepository
	Settings      repository.SettingsRepository
	Users         repository.UserRepository
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{
		db:            db,
		Pages:         repository.NewPagesRepository(db),
		Subscriptions: repository.NewSubscriptionsRepository(db),
		Preferences:   repository.NewPreferencesRepository(db),
		Notifications: repository.NewNotificationsRepository(db),
		Settings:      repository.NewSettingsRepository(db),
		Users:         repository.NewUserRepository(db),
	}
}

func (uow *UnitOfWork) DB() *gorm.DB {
	return uow.db
}

// Transaction passed a new UnitOfWork wrapped inside a gorm Transaction
func (uow *UnitOfWork) Transaction(f func(*UnitOfWork) error) error {
	return uow.db.Transaction(func(tx *gorm.DB) error {
		tempUnitOfWork := NewUnitOfWork(tx)
		return f(tempUnitOfWork)
	})
}

func (uow *UnitOfWork) Begin(opts ...*sql.TxOptions) *UnitOfWork {
	tx := uow.db.Begin(opts...)
	unitOfWork := NewUnitOfWork(tx)

	unitOfWork.isTx = true
	return unitOfWork
}

func (uow *UnitOfWork) Commit() error {
	if !uow.isTx {
		return gorm.ErrInvalidTransaction
	}

	return uow.db.Commit().Error
}

func (uow *UnitOfWork) Rollback() error {
	if !uow.isTx {
		return gorm.ErrInvalidTransaction
	}

	return uow.db.Rollback().Error
}
