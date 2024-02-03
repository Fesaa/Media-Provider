package models

import (
	"github.com/gofiber/fiber/v2"
)

type AuthProvider interface {
	// Checks the current request for authentication. This should be handled by the middleware
	IsAuthenticated(ctx *fiber.Ctx) (bool, error)

	// Logs the current user in. This happens by setting the appropriate cookie
	//
	// The request may specify a "remember me" option, which will set the cookie to expire in a month
	Login(ctx *fiber.Ctx) error

	// Logs the current user out. This happens by deleting the appropriate cookie
	Logout(ctx *fiber.Ctx) error

	// Registers a new user. This will also log them in if successful
	//
	// The request may specify a "remember me" option, which will set the cookie to expire in a month
	Register(ctx *fiber.Ctx) (*User, error)

	// Returns the currently logged in user. If no user is logged in, nil is returned.
	// While you should check for this, to prevent panics, authorization should be used to
	// check if a user is logged in.
	//
	// There is a cache for this, if it is critical that the user is up to date, use the
	// UserRaw function instead.
	User(ctx *fiber.Ctx) *User

	// Returns the currently logged in user. If no user is logged in, nil is returned.
	// This function does not use the cache, and will always query the database.
	// Only use this for critical operations, as it will be slower than User
	UserRaw(ctx *fiber.Ctx) (*User, error)
}
