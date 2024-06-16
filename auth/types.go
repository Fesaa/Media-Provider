package auth

import "github.com/gofiber/fiber/v2"

type AuthProvider interface {
	// IsAuthenticated checks the current request for authentication. This should be handled by the middleware
	IsAuthenticated(ctx *fiber.Ctx) (bool, error)

	// Login logs the current user in. This happens by setting the appropriate cookie
	//
	// The request may specify a "remember me" option, which will set the cookie to expire in a month
	Login(ctx *fiber.Ctx) error

	// Logout logs the current user out. This happens by deleting the appropriate cookie
	Logout(ctx *fiber.Ctx) error
}
