package models

type Role struct {
	id          int
	name        string
	description string
	value       int64
}

func NewRole(id int, name string, description string, value int64) Role {
	return Role{
		id:          id,
		name:        name,
		description: description,
		value:       value,
	}
}

func (r *Role) Name() string {
	return r.name
}

func (r *Role) Description() string {
	return r.description
}

func (r *Role) HasPermission(value int64) bool {
	return r.value&value == value
}
