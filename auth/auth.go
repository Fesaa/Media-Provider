package auth

import (
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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
	pass func() string
}

func newAuth() Provider {
	return &authImpl{
		pass: func() string { return config.OrDefault(config.I().Password, "admin") },
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

	token, err := jwt.ParseWithClaims(key, &MpClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(config.I().Secret), nil
	})
	if err != nil {
		return false, err
	}

	return token.Valid, nil
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

	claims := MpClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(func() time.Time {
				if body.Remember {
					return time.Now().Add(7 * 24 * time.Hour)
				}
				return time.Now().Add(24 * time.Hour)
			}()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(config.I().Secret))
	if err != nil {
		return nil, err
	}

	return &payload.LoginResponse{Token: t}, nil
}

func badRequest(msg string) error {
	return &fiber.Error{
		Code:    fiber.ErrBadRequest.Code,
		Message: msg,
	}
}
