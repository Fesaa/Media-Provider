package models

import "slices"

type User struct {
	id       int
	username string
	password string

	roles []Role
}

func NewUser(id int, username string, password string, roles []Role) *User {
	return &User{
		id:       id,
		username: username,
		password: password,
		roles:    roles,
	}
}

func (u *User) Id() int {
	return u.id
}

func (u *User) Username() string {
	return u.username
}

// nil permissions will always return false
func (u *User) HasPermission(perm *Permission) bool {
	if perm == nil {
		return false
	}
	return u.HasPermissionByValue(perm.Value())
}

func (u *User) HasPermissionByValue(value int64) bool {
	return slices.ContainsFunc(u.roles, func(r Role) bool {
		return r.HasPermission(value)
	})
}

func (u *User) HasRole(name string) bool {
	return slices.ContainsFunc(u.roles, func(r Role) bool {
		return r.Name() == name
	})
}
