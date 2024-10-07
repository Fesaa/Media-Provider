package auth

import (
	"errors"
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/gofiber/fiber/v2"
)

const (
	apiQueryKey = "api-key"
)

type apiKeyAuth struct {
}

func newApiKeyAuth() Provider {
	return apiKeyAuth{}
}

func (a apiKeyAuth) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	apiKey := ctx.Query(apiQueryKey)
	if apiKey == "" {
		return false, nil
	}
	return config.I().ApiKey == apiKey, nil
}

func (a apiKeyAuth) Login(ctx *fiber.Ctx) (*payload.LoginResponse, error) {
	log.Error("ApiKeyAuth does not support login")
	return nil, errors.New("ApiKeyAuth does not support login")
}
