package auth

import (
	"encoding/base64"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
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
	db *db.Database
}

func newJwtAuth(db *db.Database) Provider {
	return &jwtAuth{db}
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

	mpClaims, ok := token.Claims.(*MpClaims)
	if !ok {
		return false, ErrMissingOrMalformedAPIKey
	}

	// Load user from theDb in non get requests
	if ctx.Method() != fiber.MethodGet {
		user, err := jwtAuth.db.Users.GetByName(mpClaims.User.Name)
		if err != nil {
			return false, fmt.Errorf("cannot get user: %w", err)
		}
		if user == nil {
			return false, ErrMissingOrMalformedAPIKey
		}
		ctx.Locals("user", *user)
	} else {
		ctx.Locals("user", mpClaims.User)
	}

	return token.Valid, nil
}

func (jwtAuth *jwtAuth) Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	user, err := jwtAuth.db.Users.GetByName(loginRequest.UserName)
	if err != nil {
		log.Error("failed to get user by username:", "name", loginRequest.UserName)
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not found", loginRequest.UserName)
	}

	decodeString, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		log.Error("Failed to decode password, cannot login", "error", err)
		return nil, fiber.ErrInternalServerError
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(loginRequest.Password)); err != nil {
		log.Error("Invalid password, cannot login", "error", err)
		return nil, badRequest("Invalid password")
	}

	claims := MpClaims{
		User: *user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(func() time.Time {
				if loginRequest.Remember {
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
		Id:          user.ID,
		Token:       t,
		ApiKey:      user.ApiKey,
		Permissions: user.Permission,
	}, nil
}

func badRequest(msg string) error {
	return &fiber.Error{
		Code:    fiber.ErrBadRequest.Code,
		Message: msg,
	}
}
