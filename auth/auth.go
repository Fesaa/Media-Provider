package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/gofiber/fiber/v2"
	"time"
)

const (
	HeaderName = "Authorization"
	AuthScheme = "Bearer"
)

var (
	ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API key")
	authProvider                Provider
)

func Init() {
	authProvider = newAuth()
}

func I() Provider {
	return authProvider
}

type authImpl struct {
	tokens map[string]time.Time
	pass   func() string
}

func newAuth() Provider {
	return &authImpl{
		tokens: make(map[string]time.Time),
		pass:   func() string { return config.OrDefault(config.I().Password, "admin") },
	}
}

func (v *authImpl) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	auth := ctx.Get(HeaderName)
	l := len(AuthScheme)
	key, err := func() (string, error) {
		if len(auth) > 0 && l == 0 {
			return auth, nil
		}
		if len(auth) > l+1 && auth[:l] == AuthScheme {
			return auth[l+1:], nil
		}

		return "", ErrMissingOrMalformedAPIKey
	}()

	if err != nil {
		return false, err
	}
	t, ok := v.tokens[key]
	if !ok {
		return false, nil
	}

	return time.Since(t) < time.Hour*24*7, nil
}

func (v *authImpl) Login(ctx *fiber.Ctx) (*payload.LoginResponse, error) {
	body := payload.LoginRequest{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return nil, err
	}

	password := body.Password
	if password == "" {
		return nil, badRequest("Password is required")
	}

	if password != v.pass() {
		return nil, badRequest("Invalid password")
	}

	token := generateSecureToken(32)
	v.tokens[token] = time.Now().Add(time.Hour * 24 * 7)
	return &payload.LoginResponse{Token: token}, nil
}

func badRequest(msg string) error {
	return &fiber.Error{
		Code:    fiber.ErrBadRequest.Code,
		Message: msg,
	}
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
