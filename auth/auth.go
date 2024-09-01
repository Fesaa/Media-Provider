package auth

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/gofiber/fiber/v2"
	"strings"
	"time"
)

const (
	TokenCookieName = "token"
)

var authProvider Provider

func Init(cfg *config.Config) {
	authProvider = newAuth(cfg)
}

func I() Provider {
	return authProvider
}

type authImpl struct {
	cfg    *config.Config
	pass   string
	tokens map[string]time.Time
}

func newAuth(cfg *config.Config) Provider {
	return &authImpl{
		tokens: make(map[string]time.Time),
		pass:   config.OrDefault(cfg.Password, "admin"),
	}
}

func (v *authImpl) UpdatePassword(ctx *fiber.Ctx) error {
	body := payload.UpdatePasswordRequest{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return err
	}

	if body.Password == "" {
		return badRequest("Password is required")
	}

	v.pass = body.Password

	newCfg := config.I()
	newCfg.Password = v.pass
	return newCfg.Save()
}

func (v *authImpl) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	headers := ctx.GetReqHeaders()
	authorization := headers["Authorization"]
	if len(authorization) == 0 {
		return false, nil
	}
	auth := authorization[0]
	split := strings.SplitN(auth, "Bearer ", 2)
	if len(split) != 2 {
		return false, nil
	}
	token := split[1]
	if token == "" {
		return false, nil
	}

	t, ok := v.tokens[token]
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

	if password != v.pass {
		return nil, badRequest("Invalid password")
	}

	token := generateSecureToken(32)
	v.tokens[token] = time.Now().Add(time.Hour * 24 * 7)

	ctx.Cookie(&fiber.Cookie{
		Name:        TokenCookieName,
		Value:       token,
		SessionOnly: body.Remember,
		Expires:     time.Now().Add(time.Hour * 24 * 7),
	})
	return &payload.LoginResponse{Token: token}, nil
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
