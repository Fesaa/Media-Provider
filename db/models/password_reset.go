package models

import (
	"database/sql"
	"errors"
	"github.com/Fesaa/Media-Provider/utils"
	"time"
)

type PassWordReset struct {
	UserId int64
	Key    string
	Expiry time.Time
}

func (u *Users) GenerateReset(userId int64) (*PassWordReset, error) {
	key, err := utils.GenerateSecret(32)
	if err != nil {
		return nil, err
	}

	reset := &PassWordReset{
		UserId: userId,
		Key:    key,
		Expiry: time.Now().Add(time.Hour * 24),
	}

	_, err = u.db.Exec("INSERT INTO password_reset (user_id, key, expiry) VALUES (?, ?, ?)", reset.UserId, reset.Key, reset.Expiry.Unix())
	if err != nil {
		return nil, err
	}

	return reset, nil
}

func (u *Users) GetReset(key string) (*PassWordReset, error) {
	var reset PassWordReset
	var unix int64
	row := u.db.QueryRow("SELECT user_id, key, expiry FROM password_reset WHERE key = ?", key)
	err := row.Scan(&reset.UserId, &reset.Key, &unix)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	reset.Expiry = time.Unix(unix, 0)
	if reset.Expiry.Before(time.Now()) {
		return nil, nil
	}

	return &reset, nil
}

func (u *Users) DeleteReset(key string) error {
	_, err := u.db.Exec("DELETE FROM password_reset WHERE key = ?", key)
	return err
}
