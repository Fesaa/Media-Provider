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
	GroupError    NotificationGroup = "error"
)
