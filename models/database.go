package models

import (
	"database/sql"
)

type DatabaseProvider interface {
	// Returns the database pool, may be used to Query unimplemented data
	GetPool() *sql.DB

	// Fetches a user from the database by their token.
	// This loads all their roles as well
	GetUser(token string) (*User, error)
	// Tries retrieving a token from the database.
	// A nil token may be returned if the username or password is incorrect
	GetToken(username, password string) (*string, error)
	// Creates a new user in the database.
	// This also creates a new token for the user
	CreateUser(username, password string) (*User, *string, error)

	GetPermissionProvider() PermissionProvider
}
