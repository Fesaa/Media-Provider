package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

const (
	ApiQueryKey = "api-key"

	AccessTokenCookie  = "mp_access_token"
	RefreshTokenCookie = "mp_refresh_token"
	ExpiresInCookie    = "mp_expires_in"
	StateCookie        = "mp_oauth_state"

	CookieMaxAge      = 24 * 60 * 60
	StateCookieMaxAge = 10 * 60

	IdToken = "id_token"
)

type AuthMiddleware interface {
	Middleware(ctx *fiber.Ctx) error
}

type AuthService interface {
	AuthMiddleware

	Login(ctx *fiber.Ctx, loginRequest payload.LoginRequest) (*payload.LoginResponse, error)
	Logout(ctx *fiber.Ctx)
	GetOIDCLoginURL(ctx *fiber.Ctx) (string, error)
	HandleOIDCCallback(ctx *fiber.Ctx) error
}

type MpClaims struct {
	User models.User `json:"user,omitempty"`
	jwt.RegisteredClaims
}

type OIDCTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
}

var (
	ErrMissingOrMalformedAPIKey = errors.New("missing or malformed API key")
	ErrEmailNotVerified         = errors.New("email not verified")
	ErrCouldNotLinkUser         = errors.New("could not link user")
	ErrInvalidState             = errors.New("invalid OAuth state")
	ErrNoRefreshToken           = errors.New("no refresh token available")
)

type cookieAuthService struct {
	users models.Users
	cfg   *config.Config
	log   zerolog.Logger

	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier

	cookiesRefresh utils.SafeMap[uint, bool]
}

type cookieAuthServiceParams struct {
	dig.In

	Users   models.Users
	Service SettingsService
	Config  *config.Config
	Log     zerolog.Logger
}

func CookieAuthServiceProvider(params cookieAuthServiceParams) (AuthService, error) {
	settings, err := params.Service.GetSettingsDto()
	if err != nil {
		return nil, err
	}

	s := &cookieAuthService{
		users: params.Users,
		cfg:   params.Config,
		log:   params.Log.With().Str("handler", "cookie-auth-service").Logger(),
	}

	if err = s.setupOIDC(settings); err != nil {
		return nil, fmt.Errorf("failed to setup OIDC: %w", err)
	}

	return s, nil
}

func ApiKeyAuthServiceProvider(params apiKeyAuthServiceParams) AuthMiddleware {
	return &apiKeyAuthService{
		users:      params.Users,
		cookieAuth: params.FallbackAuth,
		log:        params.Log.With().Str("handler", "api-key-auth-service").Logger(),
	}
}

func (s *cookieAuthService) Login(ctx *fiber.Ctx, loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	user, err := s.users.GetByName(loginRequest.UserName)
	if err != nil {
		s.log.Error().Err(err).Str("user", loginRequest.UserName).Msg("user not found")
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not found", loginRequest.UserName)
	}

	decodeString, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		s.log.Error().Err(err).Str("user", loginRequest.UserName).Msg("failed to decode password")
		return nil, fiber.ErrInternalServerError
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(loginRequest.Password)); err != nil {
		s.log.Error().Err(err).Str("user", loginRequest.UserName).Msg("invalid password")
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
	t, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	s.setCookies(ctx, &OIDCTokens{AccessToken: t})
	return &payload.LoginResponse{
		Id:     user.ID,
		Name:   user.Name,
		Email:  user.Email.String,
		ApiKey: user.ApiKey,
		Roles:  user.Roles,
	}, nil
}

func (s *cookieAuthService) Logout(ctx *fiber.Ctx) {
	ctx.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func (s *cookieAuthService) setupOIDC(settings payload.Settings) error {
	if settings.Oidc.Authority == "" || settings.Oidc.ClientID == "" || settings.Oidc.ClientSecret == "" {
		s.log.Debug().
			Str("authority", settings.Oidc.Authority).
			Str("client_id", settings.Oidc.ClientID).
			Bool("has_secret", settings.Oidc.ClientSecret != "").
			Msg("OIDC not fully configured, skipping setup")
		return nil
	}

	provider, err := oidc.NewProvider(context.Background(), settings.Oidc.Authority)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	s.oidcProvider = provider
	s.verifier = provider.Verifier(&oidc.Config{ClientID: settings.Oidc.ClientID})

	s.oauth2Config = &oauth2.Config{
		ClientID:     settings.Oidc.ClientID,
		ClientSecret: settings.Oidc.ClientSecret,
		RedirectURL:  settings.Oidc.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	s.log.Info().
		Str("authority", settings.Oidc.Authority).
		Str("client_id", settings.Oidc.ClientID).
		Str("redirect_url", settings.Oidc.RedirectURL).
		Msg("OIDC configured successfully")
	return nil
}

func (s *cookieAuthService) Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := s.isAuthenticated(ctx)
	if !isAuthenticated {
		if err != nil {
			s.log.Debug().Err(err).Msg("error while checking authentication status")
		}
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	if err != nil {
		s.log.Debug().Err(err).Msg("error while checking authentication status")
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}

	return ctx.Next()
}

func (s *cookieAuthService) isAuthenticated(ctx *fiber.Ctx) (bool, error) {
	accessToken := ctx.Cookies(AccessTokenCookie)
	if accessToken == "" {
		return false, nil
	}

	if user, err := s.parseLocalJWT(accessToken); err == nil && user != nil {
		ctx.Locals(UserKey.Value(), *user)
		return true, nil
	}

	if s.verifier == nil {
		return false, nil
	}

	user, err := s.verifyOIDCToken(ctx.UserContext(), accessToken)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	ctx.Locals(UserKey.Value(), *user)

	expiresIn := ctx.Cookies(ExpiresInCookie)
	if expiresIn == "" {
		return true, nil
	}

	expiresSoon := utils.MustReturn(time.Parse(time.RFC3339, expiresIn)).Sub(time.Now().UTC()).Seconds() < 30
	if !expiresSoon {
		return true, nil
	}

	if isRefreshing, ok := s.cookiesRefresh.Get(user.ID); isRefreshing && ok {
		return true, nil
	}

	s.cookiesRefresh.Set(user.ID, true)
	defer s.cookiesRefresh.Delete(user.ID)

	refreshToken := ctx.Cookies(RefreshTokenCookie)
	if refreshToken == "" {
		return true, nil
	}

	newTokens, err := s.refreshOIDCToken(ctx.UserContext(), refreshToken)
	if err != nil {
		s.log.Error().Err(err).Msg("error while refreshing tokens")
		s.Logout(ctx)
		return false, err
	}

	s.log.Debug().Str("user", user.Name).Msg("refreshed tokens in background")
	s.setCookies(ctx, newTokens)
	return true, nil
}

func (s *cookieAuthService) GetOIDCLoginURL(ctx *fiber.Ctx) (string, error) {
	if s.oauth2Config == nil {
		return "", errors.New("OIDC not configured")
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	ctx.Cookie(&fiber.Cookie{
		Name:     StateCookie,
		Value:    state,
		MaxAge:   StateCookieMaxAge,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	url := s.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, nil
}

func (s *cookieAuthService) HandleOIDCCallback(ctx *fiber.Ctx) error {
	if s.oauth2Config == nil {
		return errors.New("OIDC not configured")
	}

	state := ctx.Query("state")
	storedState := ctx.Cookies(StateCookie)
	if state == "" || state != storedState {
		return ErrInvalidState
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     StateCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})

	code := ctx.Query("code")
	if code == "" {
		return errors.New("missing authorization code")
	}

	token, err := s.oauth2Config.Exchange(ctx.UserContext(), code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	rawIDToken, ok := token.Extra(IdToken).(string)
	if !ok {
		return errors.New("missing id_token")
	}

	tokens := &OIDCTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      rawIDToken,
		ExpiresIn:    int(token.Expiry.Sub(time.Now()).Seconds()),
	}

	_, err = s.verifyOIDCToken(ctx.UserContext(), rawIDToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	s.setCookies(ctx, tokens)

	return nil
}

func (s *cookieAuthService) parseLocalJWT(tokenString string) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MpClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*MpClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return &claims.User, nil
}

func (s *cookieAuthService) verifyOIDCToken(ctx context.Context, tokenString string) (*models.User, error) {
	token, err := s.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByExternalId(token.Subject)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}

	if err = token.Claims(&claims); err != nil {
		return nil, err
	}

	if !claims.Verified {
		return nil, ErrEmailNotVerified
	}

	user, err = s.users.GetByEmail(claims.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrCouldNotLinkUser
	}

	user.ExternalId = sql.NullString{String: token.Subject, Valid: true}
	if _, err = s.users.Update(*user); err != nil {
		s.log.Error().Err(err).
			Str("email", claims.Email).
			Msg("failed to assign external id to user")
		return nil, err
	}
	return user, nil
}

func (s *cookieAuthService) RefreshToken(ctx *fiber.Ctx) (*payload.LoginResponse, error) {
	refreshToken := ctx.Cookies(RefreshTokenCookie)
	if refreshToken == "" {
		return nil, ErrNoRefreshToken
	}

	tokens, err := s.refreshOIDCToken(ctx.UserContext(), refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.verifyOIDCToken(ctx.UserContext(), tokens.IDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refreshed ID token: %w", err)
	}

	s.setCookies(ctx, tokens)

	return &payload.LoginResponse{
		Id:    user.ID,
		Name:  user.Name,
		Email: user.Email.String,
		Roles: user.Roles,
	}, nil
}

func (s *cookieAuthService) refreshOIDCToken(ctx context.Context, refreshToken string) (*OIDCTokens, error) {
	if s.oauth2Config == nil {
		return nil, errors.New("OIDC not configured")
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := s.oauth2Config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	rawIDToken, ok := newToken.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("missing id_token in refresh response")
	}

	return &OIDCTokens{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		IDToken:      rawIDToken,
		ExpiresIn:    int(newToken.Expiry.Sub(time.Now()).Seconds()),
	}, nil
}

func (s *cookieAuthService) setCookies(ctx *fiber.Ctx, tokens *OIDCTokens) {
	ctx.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookie,
		Value:    tokens.AccessToken,
		MaxAge:   tokens.ExpiresIn,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	if tokens.RefreshToken != "" {
		ctx.Cookie(&fiber.Cookie{
			Name:     RefreshTokenCookie,
			Value:    tokens.RefreshToken,
			MaxAge:   CookieMaxAge * 30,
			HTTPOnly: true,
			SameSite: "Lax",
		})
	}

	if tokens.ExpiresIn != 0 {
		ctx.Cookie(&fiber.Cookie{
			Name:     ExpiresInCookie,
			Value:    time.Now().UTC().Add(time.Second * time.Duration(tokens.ExpiresIn)).Format(time.RFC3339),
			MaxAge:   CookieMaxAge * 30,
			HTTPOnly: true,
			SameSite: "Lax",
		})
	}
}

type apiKeyAuthServiceParams struct {
	dig.In

	Users        models.Users
	FallbackAuth AuthService
	Log          zerolog.Logger
}

type apiKeyAuthService struct {
	users      models.Users
	cookieAuth AuthService
	log        zerolog.Logger
}

func (a *apiKeyAuthService) Middleware(ctx *fiber.Ctx) error {
	isAuthenticated, err := a.IsAuthenticated(ctx)
	if err != nil {
		a.log.Warn().Err(err).Msg("error while checking api key auth")
	}
	if !isAuthenticated {
		return a.cookieAuth.Middleware(ctx)
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

func (a *apiKeyAuthService) Login(ctx *fiber.Ctx, loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	a.log.Error().Msg("api key auth does not support login")
	return nil, errors.New("ApiKeyAuth does not support login")
}

func (a *apiKeyAuthService) Logout(ctx *fiber.Ctx) error {
	return errors.New("ApiKeyAuth logout is not implemented")
}
