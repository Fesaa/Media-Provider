package auth

import "errors"

var (
	ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API key")
	jwtProvider                 Provider
	apiKeyProvider              Provider
)

func Init() {
	jwtProvider = newJwtAuth()
	apiKeyProvider = newApiKeyAuth()
}

func I() Provider {
	return jwtProvider
}
