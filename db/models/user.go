package models

import (
	"database/sql"
	"errors"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/log"
)

var (
	createUserStmt  *sql.Stmt
	updateUserStmt  *sql.Stmt
	deleteUserStmt  *sql.Stmt
	getUserStmt     *sql.Stmt
	getUserByApiKey *sql.Stmt
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

	getUserStmt, err = db.Prepare(`SELECT id,name,password,apiKey,permission,original FROM users WHERE name = ?`)
	if err != nil {
		return err
	}

	getUserByApiKey, err = db.Prepare(`SELECT id,name,password,apiKey,permission,original FROM users WHERE apiKey = ?`)
	if err != nil {
		return err
	}

	return nil
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

func CreateUser(name string, opts ...Option[*User]) (*User, error) {
	user := &User{Name: name}
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
	return user, nil
}

func UpdateUser(user *User, opts ...Option[*User]) (*User, error) {
	for _, opt := range opts {
		user = opt(user)
	}

	_, err := updateUserStmt.Exec(user.Name, user.PasswordHash, user.ApiKey, user.Permission, user.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func DeleteUser(id int64) error {
	_, err := deleteUserStmt.Exec(id)
	return err
}

func GetUserById(id int64) (*User, error) {
	row := db.DB.QueryRow("SELECT id,name,password,apiKey,permission,original FROM users WHERE id = ?", id)
	var user User
	if err := user.read(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func GetUser(userName string) (*User, error) {
	row := getUserStmt.QueryRow(userName)
	var user User
	if err := user.read(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}
	return &user, nil
}

func GetUserByApiKey(key string) (*User, error) {
	row := getUserByApiKey.QueryRow(key)
	var user User
	if err := user.read(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func AnyUserExists() (bool, error) {
	row := db.DB.QueryRow("SELECT COUNT(*) FROM users;")

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func Users() ([]User, error) {
	rows, err := db.DB.Query("SELECT id, name, password, apiKey, permission FROM users")
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			log.Warn("failed to close rows", "err", err)
		}
	}(rows)

	users := []User{}
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Name, &user.PasswordHash, &user.ApiKey, &user.Permission)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
