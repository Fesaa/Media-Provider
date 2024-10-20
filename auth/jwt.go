package auth

import (
	"encoding/base64"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const (
	HeaderName = "Authorization"
	AuthScheme = "Bearer"
)

type jwtAuth struct {
	pass func() string
}

func newJwtAuth() Provider {
	return &jwtAuth{
		pass: func() string { return config.I().Password },
	}
}

func (jwtAuth *jwtAuth) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
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
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}

		return []byte(config.I().Secret), nil
	})
	if err != nil {
		return false, err
	}

	return token.Valid, nil
}

func (jwtAuth *jwtAuth) Login(ctx *fiber.Ctx) (*payload.LoginResponse, error) {
	body := payload.LoginRequest{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return nil, err
	}

	password := body.Password
	if password == "" {
		return nil, badRequest("Password is required")
	}

	decodeString, err := base64.StdEncoding.DecodeString(jwtAuth.pass())
	if err != nil {
		log.Error("Failed to decode password, cannot login", "error", err, "hashed", jwtAuth.pass())
		return nil, fiber.ErrInternalServerError
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(password)); err != nil {
		log.Error("Invalid password, cannot login", "error", err)
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

	return &payload.LoginResponse{
		Token:  t,
		ApiKey: config.I().ApiKey,
	}, nil
}

func badRequest(msg string) error {
	return &fiber.Error{
		Code:    fiber.ErrBadRequest.Code,
		Message: msg,
	}
}
