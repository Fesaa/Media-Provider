package models

import (
	"database/sql"
	"github.com/Fesaa/Media-Provider/log"
	"time"
)

type Subscription struct {
	Id               int64            `json:"id"`
	Provider         Provider         `json:"provider" validator:"required"`
	ContentId        string           `json:"contentId" validator:"required"`
	RefreshFrequency RefreshFrequency `json:"refreshFrequency" validator:"required"`
	Info             Info             `json:"info" validator:"required"`
}

func (s *Subscription) read(scanner scanner) error {
	s.Info = Info{}
	return scanner.Scan(&s.Id, &s.Provider, &s.ContentId, &s.RefreshFrequency,
		&s.Info.SubscriptionId, &s.Info.Title, &s.Info.Description, &s.Info.LastCheck, &s.Info.LastCheckSuccess, &s.Info.BaseDir)
}

type RefreshFrequency int

const (
	OneHour RefreshFrequency = iota
	HalfDay
	FullDay
	Week
)

func (f RefreshFrequency) AsDuration() time.Duration {
	switch f {
	case OneHour:
		return time.Hour * 1
	case HalfDay:
		return time.Hour * 12
	case FullDay:
		return time.Hour * 24
	case Week:
		return time.Hour * 24 * 7
	}
	panic("invalid refresh frequency")
}

type Info struct {
	SubscriptionId string `json:"subscriptionId"`

	Title            string    `json:"title" validator:"required"`
	Description      string    `json:"description"`
	BaseDir          string    `json:"baseDir" validator:"required"`
	LastCheck        time.Time `json:"lastCheck"`
	LastCheckSuccess bool      `json:"lastCheckSuccess"`
}

func (i *Info) read(s scanner) error {
	return s.Scan()
}

func NewSubscriptions(db *sql.DB) *Subscriptions {
	return &Subscriptions{
		db: db,
	}
}

type Subscriptions struct {
	db *sql.DB
}

func (s *Subscriptions) All() ([]Subscription, error) {
	rows, err := s.db.Query("SELECT id, provider, contentId, refreshFrequency, subscription_id, title, description, lastCheck, lastCheckSuccess, baseDir FROM subscriptions s JOIN subscription_info si ON s.id = si.subscription_id")

	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		if err = sub.read(rows); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

func (s *Subscriptions) Get(i int64) (*Subscription, error) {
	row := s.db.QueryRow(`SELECT id, provider, contentId, refreshFrequency,
       subscription_id, title, description, lastCheck, lastCheckSuccess, baseDir
		FROM subscriptions s JOIN subscription_info si ON s.id = si.subscription_id WHERE s.id = $1`, i)

	var sub Subscription
	if err := sub.read(row); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *Subscriptions) Update(subscriptions ...*Subscription) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		if err = upsert(tx, sub); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Error("failed to rollback transaction", "err", err, "rollbackErr", rollbackErr)
			}
			return err
		}
	}

	return tx.Commit()
}

func (s *Subscriptions) New(model Subscription) (*Subscription, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec("INSERT INTO subscriptions (provider, contentId, refreshFrequency) VALUES ($1, $2, $3)", model.Provider, model.ContentId, model.RefreshFrequency)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			log.Warn("failed to rollback transaction", "err", err)
		}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec("INSERT INTO subscription_info (subscription_id, title, description, baseDir, lastCheck, lastCheckSuccess) VALUES ($1, $2, $3, $4, date('now'), 1)", id, model.Info.Title, model.Info.Description, model.Info.BaseDir)
	if err != nil {
		if err = tx.Rollback(); err != nil {
			log.Warn("failed to rollback transaction", "err", err)
		}
		return nil, err
	}

	model.Id = id
	return &model, tx.Commit()
}

func upsert(tx *sql.Tx, s *Subscription) error {
	_, err := tx.Exec("UPDATE subscriptions SET provider = $1, contentId = $2, refreshFrequency = $3 WHERE id = $4;", s.Provider, s.ContentId, s.RefreshFrequency, s.Id)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE subscription_info SET title = $1, description = $2, baseDir = $4 WHERE subscription_id = $3;", s.Info.Title, s.Info.Description, s.Id, s.Info.BaseDir)
	return err
}

func (s *Subscriptions) Delete(i int64) error {
	_, err := s.db.Exec(`DELETE FROM subscription_info WHERE subscription_id = $1; DELETE FROM subscriptions WHERE id = $1;`, i)
	return err
}
