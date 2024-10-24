package auth

import (
	"errors"
	"github.com/Fesaa/Media-Provider/db"
)

var (
	ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API key")
	jwtProvider                 Provider
	apiKeyProvider              Provider
)

func Init(db *db.Database) {
	jwtProvider = newJwtAuth(db)
	apiKeyProvider = newApiKeyAuth(db)
}

func I() Provider {
	return jwtProvider
}
