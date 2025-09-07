package db

import (
	"github.com/Fesaa/Media-Provider/db/repository"
	"github.com/devfeel/mapper"
	"gorm.io/gorm"
)

type UnitOfWork struct {
	db *gorm.DB
	tx *gorm.DB
	m  mapper.IMapper

	Pages         repository.PagesRepository
	Subscriptions repository.SubscriptionsRepository
	Preferences   repository.PreferencesRepository
	Notifications repository.NotificationsRepository
	Settings      repository.SettingsRepository
	Users         repository.UserRepository
}

func NewUnitOfWork(db *gorm.DB, m mapper.IMapper) *UnitOfWork {
	tx := db.Begin()
	return &UnitOfWork{
		db:            db,
		tx:            tx,
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

/**
TODO: This is part of the more complex rewrite towards a more .NET approach
	And safer DB handling

// Commit commits the current transaction, then starts a new one
func (uow *UnitOfWork) Commit() error {
	if err := uow.tx.Commit().Error; err != nil {
		return err
	}

	return uow.begin()
}

// Rollback rolls the current transaction back, then starts a new one
func (uow *UnitOfWork) Rollback() error {
	if err := uow.tx.Rollback().Error; err != nil {
		return err
	}

	return uow.begin()
}

// Begin starts a new transaction
func (uow *UnitOfWork) begin() error {
	tx := uow.db.Begin()
	m := uow.m

	uow.tx = tx
	uow.Pages = repository.NewPagesRepository(tx, m)
	uow.Subscriptions = repository.NewSubscriptionsRepository(tx, m)
	uow.Preferences = repository.NewPreferencesRepository(tx, m)
	uow.Notifications = repository.NewNotificationsRepository(tx, m)
	uow.Settings = repository.NewSettingsRepository(tx, m)
	uow.Users = repository.NewUserRepository(tx, m)
	return nil
}

// Close rolls back the current transaction if it's still open
func (uow *UnitOfWork) Close() error {
	if uow.tx != nil {
		return uow.tx.Rollback().Error
	}
	return nil
}
**/
