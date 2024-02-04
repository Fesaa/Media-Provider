package impl

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	TokenCookieName = "token"
)

type AuthImpl struct {
	pass   string
	tokens map[string]time.Time
}

func newAuth() *AuthImpl {
	return &AuthImpl{
		pass:   getEnv("PASS", "admin"),
		tokens: make(map[string]time.Time),
	}
}

func getEnv(key string, defaultValue ...string) string {
	e, b := os.LookupEnv(key)
	if !b && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return e
}

func (v *AuthImpl) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	token := ctx.Cookies(TokenCookieName)
	if token == "" {
		return false, nil
	}

	_, ok := v.tokens[token]
	return ok, nil
}

type LoginBody struct {
	Password string `json:"password"`
	Remember string `json:"remember,omitempty"`
}

func (v *AuthImpl) Login(ctx *fiber.Ctx) error {
	body := LoginBody{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return err
	}

	password := body.Password
	if password == "" {
		return badRequest("Password is required")
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

func (v *AuthImpl) Logout(ctx *fiber.Ctx) error {
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
