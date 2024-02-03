package models

type PermissionProvider interface {
	// Return all loaded permissions. Is never nil
	GetAllPermissions() []*Permission

	// Return the permission with the given key or nil if not found
	GetPermissionByKey(key string) *Permission

	// Check if a user has a permission by key
	UserHasPermission(u *User, key string) bool

	// Check if a role has a permission by key
	RoleHasPermission(r *Role, key string) bool

	// Load all permissions from the database
	RefreshPermissions() error
}
