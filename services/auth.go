package services

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"golang.org/x/crypto/bcrypt"
)

const (
	ApiQueryKey = "api-key"
	Header      = "Authorization"
	Scheme      = "Bearer"
)

type AuthService interface {
	// IsAuthenticated checks the current request for authentication. This should be handled by the middleware
	IsAuthenticated(ctx *fiber.Ctx) (bool, error)

	// Login logs the current user in.
	Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error)

	Middleware(ctx *fiber.Ctx) error
}

type MpClaims struct {
	User models.User `json:"user,omitempty"`
	jwt.RegisteredClaims
}

var (
	ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API key")
	ErrEmailNotVerified         = errors.New("email not verified")
	ErrCouldNotLinkUser         = errors.New("could not link user")
)

type jwtAuthService struct {
	users models.Users
	cfg   *config.Config
	log   zerolog.Logger

	iss      string
	verifier *oidc.IDTokenVerifier
}

func JwtAuthServiceProvider(service SettingsService, users models.Users, cfg *config.Config, log zerolog.Logger) (AuthService, error) {
	settings, err := service.GetSettingsDto()
	if err != nil {
		return nil, err
	}

	s := &jwtAuthService{
		users: users,
		cfg:   cfg,
		log:   log.With().Str("handler", "jwt-auth-service").Logger(),
	}

	verifier, err := s.oidcTokenVerifier(settings)
	if err != nil {
		return nil, err
	}

	s.verifier = verifier
	s.iss = settings.Oidc.Authority
	return s, nil
}

func ApiKeyAuthServiceProvider(params apiKeyAuthServiceParams) AuthService {
	return &apiKeyAuthService{
		users: params.Users,
		jwt:   params.JWT,
		log:   params.Log.With().Str("handler", "api-key-auth-service").Logger(),
	}
}

func GetIssuerFromToken(tokenString string) (string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	issuer, ok := claims["iss"].(string)
	if !ok {
		return "", errors.New("issuer claim not found or not a string")
	}

	return issuer, nil
}

func (jwtAuth *jwtAuthService) Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := jwtAuth.IsAuthenticated(ctx)
	if !isAuthenticated {
		if err != nil {
			jwtAuth.log.Debug().Err(err).Msg("error while checking authentication status")
		}

		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	if err != nil {
		jwtAuth.log.Debug().Err(err).Msg("error while checking authentication status")
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	return ctx.Next()
}

func (jwtAuth *jwtAuthService) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	auth := ctx.Get(Header)
	l := len(Scheme)
	key, err := func() (string, error) {
		if len(auth) > 0 && l == 0 {
			return auth, nil
		}
		if len(auth) > l+1 && auth[:l] == Scheme {
			return auth[l+1:], nil
		}

		return "", ErrMissingOrMalformedAPIKey
	}()

	if err != nil {
		return false, err
	}

	iss, err := GetIssuerFromToken(key)
	if err != nil {
		return false, err
	}

	if jwtAuth.verifier != nil && iss == jwtAuth.iss {
		return jwtAuth.OidcJWT(ctx, key)
	}

	return jwtAuth.LocalJWT(ctx, key)
}

func (jwtAuth *jwtAuthService) OidcJWT(ctx *fiber.Ctx, key string) (bool, error) {
	token, err := jwtAuth.verifier.Verify(ctx.UserContext(), key)
	if err != nil {
		return false, err
	}

	user, err := jwtAuth.users.GetByExternalId(token.Subject)
	if err != nil {
		return false, err
	}

	if user != nil {
		ctx.Locals(UserKey.Value(), *user)
		return true, nil
	}

	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}
	if err = token.Claims(&claims); err != nil {
		return false, err
	}

	if !claims.Verified {
		return false, ErrEmailNotVerified
	}

	user, err = jwtAuth.users.GetByEmail(claims.Email)
	if err != nil {
		return false, err
	}
	if user != nil {
		user.ExternalId = sql.NullString{
			String: token.Subject,
			Valid:  true,
		}
		if _, err = jwtAuth.users.Update(*user); err != nil {
			jwtAuth.log.Error().Err(err).
				Str("email", claims.Email).
				Msg("failed to assign external id to user")
			return false, err
		}

		ctx.Locals(UserKey.Value(), *user)
		return true, nil
	}

	return false, ErrCouldNotLinkUser
}

func (jwtAuth *jwtAuthService) LocalJWT(ctx *fiber.Ctx, key string) (bool, error) {
	token, err := jwt.ParseWithClaims(key, &MpClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}

		return []byte(jwtAuth.cfg.Secret), nil
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
		user, err := jwtAuth.users.GetById(mpClaims.User.ID)
		if err != nil {
			return false, fmt.Errorf("cannot get user: %w", err)
		}
		if user == nil {
			return false, ErrMissingOrMalformedAPIKey
		}
		ctx.Locals(UserKey.Value(), *user)
	} else {
		ctx.Locals(UserKey.Value(), mpClaims.User)
	}

	return token.Valid, nil
}

func (jwtAuth *jwtAuthService) Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	user, err := jwtAuth.users.GetByName(loginRequest.UserName)
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
		return nil, &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "invalid password",
		}
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
			Issuer: "Media-Provider",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(jwtAuth.cfg.Secret))
	if err != nil {
		return nil, err
	}

	return &payload.LoginResponse{
		Id:     user.ID,
		Name:   user.Name,
		Email:  user.Email.String,
		Token:  t,
		ApiKey: user.ApiKey,
		Roles:  user.Roles,
	}, nil
}

func (jwtAuth *jwtAuthService) oidcTokenVerifier(dto payload.Settings) (*oidc.IDTokenVerifier, error) {
	if dto.Oidc.Authority == "" || dto.Oidc.ClientID == "" {
		jwtAuth.log.Debug().
			Str("authority", dto.Oidc.Authority).
			Str("client_id", dto.Oidc.ClientID).
			Msg("not setting up OIDC")
		return nil, nil
	}

	provider, err := oidc.NewProvider(context.Background(), dto.Oidc.Authority)
	if err != nil {
		return nil, err
	}

	return provider.Verifier(&oidc.Config{ClientID: dto.Oidc.ClientID}), nil
}

type apiKeyAuthServiceParams struct {
	dig.In

	Users models.Users
	JWT   AuthService `name:"jwt-auth"`
	Log   zerolog.Logger
}

type apiKeyAuthService struct {
	users models.Users
	jwt   AuthService
	log   zerolog.Logger
}

func (a *apiKeyAuthService) Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := a.IsAuthenticated(ctx)
	if err != nil {
		a.log.Warn().Err(err).Msg("error while checking api key auth")
	}
	if !isAuthenticated {
		return a.jwt.Middleware(ctx)
	}
	return ctx.Next()
}

func (a *apiKeyAuthService) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	apiKey := ctx.Query(ApiQueryKey)
	if apiKey == "" {
		return false, nil
	}

	user, err := a.users.GetByApiKey(apiKey)
	if err != nil {
		return false, err
	}

	ctx.Locals(UserKey.Value(), user)
	return true, nil
}

func (a *apiKeyAuthService) Login(loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	a.log.Error().Msg("api key auth does not support login")
	return nil, errors.New("ApiKeyAuth does not support login")
}
