package models

import (
	"database/sql"
	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model

	Title   string             `json:"title"`
	Summary string             `json:"summary"`
	Body    string             `json:"body"`
	Colour  NotificationColour `json:"colour"`
	Group   NotificationGroup  `json:"group" gorm:"index"`

	Read   bool         `json:"read"`
	ReadAt sql.NullTime `json:"readAt"`
}

type NotificationColour string

// These are PrimeNG colours https://primeng.org/button#severity
const (
	Black    NotificationColour = "primary"
	White    NotificationColour = "secondary"
	Green    NotificationColour = "success"
	Blue     NotificationColour = "info"
	Orange   NotificationColour = "warn"
	Purple   NotificationColour = "help"
	Red      NotificationColour = "danger"
	Contrast NotificationColour = "contrast"
)

type NotificationGroup string

const (
	GroupContent  NotificationGroup = "content"
	GroupSecurity NotificationGroup = "security"
	GroupGeneral  NotificationGroup = "general"
)
