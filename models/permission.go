package models

type Permission struct {
	key         string
	description string
	value       int64
}

func NewPermission(key string, description string, value int64) *Permission {
	return &Permission{
		key:         key,
		description: description,
		value:       value,
	}
}

func (p *Permission) Key() string {
	return p.key
}

func (p *Permission) Description() string {
	return p.description
}

func (p *Permission) Value() int64 {
	return p.value
}
