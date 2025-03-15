package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const (
	HeaderName = "Authorization"
	AuthScheme = "Bearer"
)

var (
	ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API key")
)

type jwtAuth struct {
	DB  *db.Database
	Cfg *config.Config
	log zerolog.Logger
}

func NewJwtAuth(db *db.Database, cfg *config.Config, log zerolog.Logger) Provider {
	return &jwtAuth{db, cfg,
		log.With().Str("handler", "jwt-auth").Logger(),
	}
}

func (jwtAuth *jwtAuth) Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := jwtAuth.IsAuthenticated(ctx)
	if err != nil {
		jwtAuth.log.Debug().Err(err).Msg("error while checking authentication status")
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	if !isAuthenticated {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	return ctx.Next()
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

		return []byte(jwtAuth.Cfg.Secret), nil
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
		user, err := jwtAuth.DB.Users.GetById(mpClaims.User.ID)
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
	user, err := jwtAuth.DB.Users.GetByName(loginRequest.UserName)
	if err != nil {
		jwtAuth.log.Error().Err(err).Str("user", loginRequest.UserName).Msg("user not found")
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not found", loginRequest.UserName)
	}

	decodeString, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		jwtAuth.log.Error().Err(err).Str("user", loginRequest.UserName).Msg("failed to decode password")
		return nil, fiber.ErrInternalServerError
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(loginRequest.Password)); err != nil {
		jwtAuth.log.Error().Err(err).Str("user", loginRequest.UserName).Msg("invalid password")
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
	t, err := token.SignedString([]byte(jwtAuth.Cfg.Secret))
	if err != nil {
		return nil, err
	}

	return &payload.LoginResponse{
		Id:          user.ID,
		Name:        user.Name,
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
