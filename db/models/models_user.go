package models

import (
	"database/sql"
	"encoding/json"
	"slices"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Role string

type Roles []Role

func (r Roles) HasRole(role Role) bool {
	return slices.Contains(r, role)
}

func (u *User) HasRole(role Role) bool {
	return u.Roles.HasRole(role)
}

const (
	ManagePages         Role = "manage-pages"
	ManageUsers         Role = "manage-users"
	ManageServerConfigs Role = "manage-server-configs"
	ManagePreferences   Role = "manage-preferences"
	ManageSubscriptions Role = "manage-subscriptions"
	ViewAllDownloads    Role = "view-all-downloads"
)

var AllRoles = []Role{
	ManagePages,
	ManageUsers,
	ManageServerConfigs,
	ManagePreferences,
	ManageSubscriptions,
	ViewAllDownloads,
}

type User struct {
	Model
	Name         string         `gorm:"unique"`
	Email        sql.NullString `gorm:"unique"`
	PasswordHash string
	ApiKey       string
	// Will not be updated in the UpdateUser method, should be set on creation. And only for the first account
	Original   bool
	ExternalId sql.NullString  `gorm:"unique,nullable"`
	SqlRoles   json.RawMessage `gorm:"roles"`
	Roles      Roles           `gorm:"-"`
	Pages      pq.Int32Array   `gorm:"type:integer[]"`
}

func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	u.SqlRoles, err = json.Marshal(u.Roles)
	return
}

func (u *User) AfterFind(tx *gorm.DB) (err error) {
	if u.Pages == nil {
		u.Pages = []int32{}
	}

	if u.SqlRoles == nil {
		u.Roles = []Role{}
		return
	}

	return json.Unmarshal(u.SqlRoles, &u.Roles)
}

type PasswordReset struct {
	Model

	UserId int
	Key    string
	Expiry time.Time
}
