package models

import (
	"database/sql"
	"errors"
	"github.com/Fesaa/Media-Provider/log"
)

var (
	createUserStmt *sql.Stmt
	updateUserStmt *sql.Stmt
	deleteUserStmt *sql.Stmt
)

func initUser(db *sql.DB) error {
	var err error

	createUserStmt, err = db.Prepare(`INSERT INTO users (name, password, apiKey,permission,original) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	updateUserStmt, err = db.Prepare(`UPDATE users SET name = ?, password = ?, apiKey = ?, permission = ? WHERE id = ?`)
	if err != nil {
		return err
	}

	deleteUserStmt, err = db.Prepare(`DELETE FROM users WHERE id = ?`)
	if err != nil {
		return err
	}

	return nil
}

func NewUsers(db *sql.DB) *Users {
	return &Users{
		db: db,
	}
}

type Users struct {
	db *sql.DB
}

func (u *Users) All() ([]User, error) {
	rows, err := u.db.Query("SELECT id, name, password, apiKey, permission, original FROM users")
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	out := make([]User, 0)
	for rows.Next() {
		var user User
		if err = user.read(rows); err != nil {
			return nil, err
		}
		out = append(out, user)
	}

	return out, nil
}

func (u *Users) ExistsAny() (bool, error) {
	row := u.db.QueryRow("SELECT COUNT(*) FROM users;")
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (u *Users) intoUser(s scanner) (*User, error) {
	var user User
	if err := user.read(s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	return &user, nil
}

func (u *Users) GetById(id int64) (*User, error) {
	row := u.db.QueryRow("SELECT id, name, password, apiKey, permission,original FROM users WHERE id = ?", id)
	return u.intoUser(row)
}

func (u *Users) GetByName(name string) (*User, error) {
	row := u.db.QueryRow("SELECT id, name, password, apiKey, permission,original FROM users WHERE name = ?", name)
	return u.intoUser(row)
}

func (u *Users) GetByApiKey(key string) (*User, error) {
	row := u.db.QueryRow("SELECT id, name, password, apiKey, permission,original FROM users WHERE apiKey = ?", key)
	return u.intoUser(row)
}

func (u *Users) Create(name string, opts ...Option[User]) (*User, error) {
	user := User{Name: name}
	for _, opt := range opts {
		user = opt(user)
	}

	result, err := createUserStmt.Exec(user.Name, user.PasswordHash, user.ApiKey, user.Permission, user.Original)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	user.ID = id
	return &user, nil
}

func (u *Users) Update(user User, opts ...Option[User]) (*User, error) {
	for _, opt := range opts {
		user = opt(user)
	}

	_, err := updateUserStmt.Exec(user.Name, user.PasswordHash, user.ApiKey, user.Permission, user.ID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *Users) UpdateById(id int64, opts ...Option[User]) (*User, error) {
	user, err := u.GetById(id)
	if err != nil || user == nil {
		return user, err
	}
	return u.Update(*user, opts...)
}

func (u *Users) Delete(id int64) error {
	_, err := deleteUserStmt.Exec(id)
	return err
}

type UserPermission int

const (
	PermWritePage = 1 << iota
	PermDeletePage

	PermWriteUser
	PermDeleteUser

	PermWriteConfig
)

var (
	ALL_PERMS = PermWritePage |
		PermDeletePage |
		PermWriteConfig |
		PermWriteUser |
		PermDeleteUser |
		PermWriteConfig
)

type User struct {
	ID           int64
	Name         string
	PasswordHash string
	ApiKey       string
	Permission   int
	// Will not be updated in the UpdateUser method, should be set on creation. And only for the first account
	Original bool
}

func (u *User) read(s scanner) error {
	return s.Scan(&u.ID, &u.Name, &u.PasswordHash, &u.ApiKey, &u.Permission, &u.Original)
}

func (u *User) HasPermission(permission UserPermission) bool {
	return u.Permission&int(permission) == int(permission)
}
