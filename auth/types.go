package auth

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type MpClaims struct {
	User models.User `json:"user,omitempty"`
	jwt.RegisteredClaims
}

type Provider interface {
	// IsAuthenticated checks the current request for authentication. This should be handled by the middleware
	IsAuthenticated(ctx *fiber.Ctx) (bool, error)

	// Login logs the current user in.
	Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error)
}
