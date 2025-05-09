package models

import (
	"testing"
)

func TestUser_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		user       User
		permission UserPermission
		expected   bool
	}{
		{
			name:       "Has WritePage permission",
			user:       User{Permission: int(PermWritePage)},
			permission: PermWritePage,
			expected:   true,
		},
		{
			name:       "Has DeletePage permission",
			user:       User{Permission: int(PermDeletePage)},
			permission: PermDeletePage,
			expected:   true,
		},
		{
			name:       "Has WriteUser permission",
			user:       User{Permission: int(PermWriteUser)},
			permission: PermWriteUser,
			expected:   true,
		},
		{
			name:       "Has DeleteUser permission",
			user:       User{Permission: int(PermDeleteUser)},
			permission: PermDeleteUser,
			expected:   true,
		},
		{
			name:       "Has WriteConfig permission",
			user:       User{Permission: int(PermWriteConfig)},
			permission: PermWriteConfig,
			expected:   true,
		},
		{
			name:       "Has multiple permissions",
			user:       User{Permission: int(PermWritePage | PermDeleteUser)},
			permission: PermWritePage,
			expected:   true,
		},
		{
			name:       "Has multiple permissions, but not the one we're checking",
			user:       User{Permission: int(PermWritePage | PermDeleteUser)},
			permission: PermWriteConfig,
			expected:   false,
		},
		{
			name:       "Has no permissions",
			user:       User{Permission: 0},
			permission: PermWritePage,
			expected:   false,
		},
		{
			name:       "Has all permissions",
			user:       User{Permission: ALL_PERMS},
			permission: PermWritePage,
			expected:   true,
		},
		{
			name:       "Has all permissions",
			user:       User{Permission: ALL_PERMS},
			permission: PermDeleteUser,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.HasPermission(tt.permission)
			if result != tt.expected {
				t.Errorf("HasPermission(%v) = %v, expected %v", tt.permission, result, tt.expected)
			}
		})
	}
}
