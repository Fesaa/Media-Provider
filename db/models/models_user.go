package models

import (
	"time"
)

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
	Model
	Name         string `gorm:"unique"`
	Email        string `gorm:"unique"`
	PasswordHash string
	ApiKey       string
	Permission   int
	// Will not be updated in the UpdateUser method, should be set on creation. And only for the first account
	Original   bool
	ExternalId string `gorm:"unique"`
}

func (u *User) HasPermission(permission UserPermission) bool {
	return u.Permission&int(permission) == int(permission)
}

type PasswordReset struct {
	Model

	UserId uint
	Key    string
	Expiry time.Time
}
