package db

import (
	"database/sql"

	"github.com/Fesaa/Media-Provider/db/repository"
	"github.com/devfeel/mapper"
	"gorm.io/gorm"
)

type UnitOfWork struct {
	db *gorm.DB
	m  mapper.IMapper

	isTx bool

	Pages         repository.PagesRepository
	Subscriptions repository.SubscriptionsRepository
	Preferences   repository.PreferencesRepository
	Notifications repository.NotificationsRepository
	Settings      repository.SettingsRepository
	Users         repository.UserRepository
}

func NewUnitOfWork(db *gorm.DB, m mapper.IMapper) *UnitOfWork {
	return &UnitOfWork{
		db:            db,
		m:             m,
		Pages:         repository.NewPagesRepository(db, m),
		Subscriptions: repository.NewSubscriptionsRepository(db, m),
		Preferences:   repository.NewPreferencesRepository(db, m),
		Notifications: repository.NewNotificationsRepository(db, m),
		Settings:      repository.NewSettingsRepository(db, m),
		Users:         repository.NewUserRepository(db, m),
	}
}

func (uow *UnitOfWork) DB() *gorm.DB {
	return uow.db
}

// Transaction passed a new UnitOfWork wrapped inside a gorm Transaction
func (uow *UnitOfWork) Transaction(f func(*UnitOfWork) error) error {
	return uow.db.Transaction(func(tx *gorm.DB) error {
		tempUnitOfWork := NewUnitOfWork(tx, uow.m)
		return f(tempUnitOfWork)
	})
}

func (uow *UnitOfWork) Begin(opts ...*sql.TxOptions) *UnitOfWork {
	tx := uow.db.Begin(opts...)
	unitOfWork := NewUnitOfWork(tx, uow.m)

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
