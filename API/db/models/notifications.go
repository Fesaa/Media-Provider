package models

import (
	"database/sql"

	"github.com/Fesaa/Media-Provider/utils"
	"github.com/lib/pq"
)

type Notification struct {
	Model

	Title         string             `json:"title"`
	Summary       string             `json:"summary"`
	Body          string             `json:"body"`
	Colour        NotificationColour `json:"colour"`
	Group         NotificationGroup  `json:"group" gorm:"index"`
	Owner         sql.NullInt32      `json:"owner"`
	RequiredRoles pq.StringArray     `json:"required_roles" gorm:"type:text[]"`
	Read          bool               `json:"read"`
}

func (n Notification) HasAccess(user User) bool {
	if n.Owner.Valid && n.Owner.Int32 == int32(user.ID) {
		return true
	}

	if len(n.RequiredRoles) == 0 {
		return true
	}

	return utils.Contains(n.RequiredRoles, utils.MapToString(user.Roles))
}

type NotificationBuilder struct {
	title         string
	summary       string
	body          string
	colour        NotificationColour
	group         NotificationGroup
	owner         int
	requiredRoles Roles
	read          bool
}

func NewNotification() NotificationBuilder {
	return NotificationBuilder{}
}

func (n NotificationBuilder) WithTitle(title string) NotificationBuilder {
	n.title = title
	return n
}

func (n NotificationBuilder) WithSummary(summary string) NotificationBuilder {
	n.summary = summary
	return n
}

// WithBody if no summary is set, the body will be used to create it with utils.Shorten
func (n NotificationBuilder) WithBody(body string) NotificationBuilder {
	n.body = body
	return n
}

func (n NotificationBuilder) WithColour(colour NotificationColour) NotificationBuilder {
	n.colour = colour
	return n
}

func (n NotificationBuilder) WithGroup(group NotificationGroup) NotificationBuilder {
	n.group = group
	return n
}

func (n NotificationBuilder) WithOwner(owner int) NotificationBuilder {
	n.owner = owner
	return n
}

func (n NotificationBuilder) WithRequiredRoles(requiredRoles ...Role) NotificationBuilder {
	n.requiredRoles = append(n.requiredRoles, requiredRoles...)
	return n
}

func (n NotificationBuilder) WithRead(read bool) NotificationBuilder {
	n.read = read
	return n
}

func (n NotificationBuilder) Build() Notification {
	if n.summary == "" && n.body != "" {
		n.summary = utils.Shorten(n.body, 100)
	}

	return Notification{
		Title:         n.title,
		Summary:       n.summary,
		Body:          n.body,
		Colour:        n.colour,
		Group:         n.group,
		Owner:         sql.NullInt32{Int32: int32(n.owner), Valid: n.owner != 0},
		RequiredRoles: utils.MapToString(n.requiredRoles),
		Read:          n.read,
	}
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
