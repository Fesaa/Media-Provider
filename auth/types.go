package auth

import (
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type MpClaims struct {
	jwt.RegisteredClaims
}

type Provider interface {
	// IsAuthenticated checks the current request for authentication. This should be handled by the middleware
	IsAuthenticated(ctx *fiber.Ctx) (bool, error)

	// Login logs the current user in. This happens by setting the appropriate cookie
	Login(ctx *fiber.Ctx) (*payload.LoginResponse, error)
}
