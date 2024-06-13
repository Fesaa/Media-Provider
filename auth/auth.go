package auth

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
	"time"
)

const (
	TokenCookieName = "token"
)

var authProvider AuthProvider

func Init() {
	authProvider = newAuth()
}

func I() AuthProvider {
	return authProvider
}

type authImpl struct {
	pass   string
	tokens map[string]time.Time
}

func newAuth() AuthProvider {
	return &authImpl{
		tokens: make(map[string]time.Time),
		pass:   config.OrDefault(config.I().GetPassWord(), "admin"),
	}
}

func (v *authImpl) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	token := ctx.Cookies(TokenCookieName)
	if token == "" {
		return false, nil
	}

	t, ok := v.tokens[token]
	if !ok {
		return false, nil
	}

	return time.Since(t) < time.Hour*24*7, nil
}

func (v *authImpl) Login(ctx *fiber.Ctx) error {
	body := LoginRequest{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return err
	}

	password := body.Password
	if password == "" {
		return badRequest("Password is required")
	}

	if password != v.pass {
		return badRequest("Invalid password")
	}

	sessionOnly := body.Remember == ""

	token := generateSecureToken(32)
	v.tokens[token] = time.Now().Add(time.Hour * 24 * 7)

	ctx.Cookie(&fiber.Cookie{
		Name:        TokenCookieName,
		Value:       token,
		SessionOnly: sessionOnly,
		Expires:     time.Now().Add(time.Hour * 24 * 7),
	})
	return nil
}

func (v *authImpl) Logout(ctx *fiber.Ctx) error {
	ctx.Cookie(&fiber.Cookie{
		Name:    TokenCookieName,
		Expires: time.Now().Add(-(time.Hour * 5)),
	})
	return nil
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
