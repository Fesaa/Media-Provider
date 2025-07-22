package models

type Notification struct {
	Model

	Title   string             `json:"title"`
	Summary string             `json:"summary"`
	Body    string             `json:"body"`
	Colour  NotificationColour `json:"colour"`
	Group   NotificationGroup  `json:"group" gorm:"index"`

	Read bool `json:"read"`
}

type NotificationColour string

const (
	Primary   NotificationColour = "primary"
	Secondary NotificationColour = "secondary"
	Warning   NotificationColour = "warning"
	Error     NotificationColour = "error"
)

type NotificationGroup string

const (
	GroupContent  NotificationGroup = "content"
	GroupSecurity NotificationGroup = "security"
	GroupGeneral  NotificationGroup = "general"
	GroupError    NotificationGroup = "error"
)
