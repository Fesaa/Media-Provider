package auth

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
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

	user, err := models.GetUserByApiKey(apiKey)
	if err != nil {
		return false, err
	}

	return user.ApiKey == apiKey, nil
}

func (a apiKeyAuth) Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	log.Error("ApiKeyAuth does not support login")
	return nil, errors.New("ApiKeyAuth does not support login")
}
