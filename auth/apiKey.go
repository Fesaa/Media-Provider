package auth

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

const (
	apiQueryKey = "api-key"
)

type apiKeyAuth struct {
	db  *db.Database
	jwt Provider
	log zerolog.Logger
}

type apiKeyAuthParams struct {
	dig.In

	DB  *db.Database
	JWT Provider `name:"jwt-auth"`
	Log zerolog.Logger
}

func NewApiKeyAuth(params apiKeyAuthParams) Provider {
	return apiKeyAuth{params.DB, params.JWT,
		params.Log.With().Str("handler", "api-key-auth").Logger(),
	}
}

func (a apiKeyAuth) Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := a.IsAuthenticated(ctx)
	if err != nil {
		a.log.Warn().Err(err).Msg("error while checking api key auth")
	}
	if !isAuthenticated {
		return a.jwt.Middleware(ctx)
	}
	return ctx.Next()
}

func (a apiKeyAuth) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	apiKey := ctx.Query(apiQueryKey)
	if apiKey == "" {
		return false, nil
	}

	user, err := a.db.Users.GetByApiKey(apiKey)
	if err != nil {
		return false, err
	}

	ctx.Locals("user", user)
	return true, nil
}

func (a apiKeyAuth) Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	a.log.Error().Msg("api key auth does not support login")
	return nil, errors.New("ApiKeyAuth does not support login")
}
