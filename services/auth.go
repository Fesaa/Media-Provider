package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
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
	apiQueryKey = "api-key"

	tokenCookie = "mp_token"
	stateCookie = "mp_oauth_state"
	idToken     = "id_token"

	stateCookieMaxAge = 10 * 60
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

type mpClaims struct {
	User models.User `json:"user,omitempty"`
	jwt.RegisteredClaims
}

type OIDCTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token"`
	ExpiresIn    time.Time `json:"expires_in"`
}

var (
	errEmailNotVerified = errors.New("email not verified")
	errCouldNotLinkUser = errors.New("could not link user")
	errInvalidState     = errors.New("invalid OAuth state")
)

type cookieAuthService struct {
	users   models.Users
	cfg     *config.Config
	log     zerolog.Logger
	storage CacheService

	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier

	cookiesRefresh utils.SafeMap[uint, bool]
}

type cookieAuthServiceParams struct {
	dig.In

	Users   models.Users
	Service SettingsService
	Storage CacheService
	Config  *config.Config
	Log     zerolog.Logger
}

func CookieAuthServiceProvider(params cookieAuthServiceParams) (AuthService, error) {
	settings, err := params.Service.GetSettingsDto()
	if err != nil {
		return nil, err
	}

	s := &cookieAuthService{
		users:   params.Users,
		storage: params.Storage,
		cfg:     params.Config,
		log:     params.Log.With().Str("handler", "cookie-auth-service").Logger(),
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

	claims := mpClaims{
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

	if err = s.setCookies(ctx, &OIDCTokens{AccessToken: t}); err != nil {
		return nil, err
	}

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
		Name:     tokenCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     stateCookie,
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
	token := ctx.Cookies(tokenCookie)
	if token == "" {
		return false, nil
	}

	tokens, err := s.getOidcTokens(token)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			s.log.Warn().Msg("consider setting up a redis store, to ensure authentication survives restarts.")
		}
		return false, fmt.Errorf("failed to get tokens from storage: %w", err)
	}

	if user, err := s.parseLocalJWT(tokens.AccessToken); err == nil && user != nil {
		ctx.Locals(UserKey.Value(), *user)
		return true, nil
	}

	if s.verifier == nil {
		return false, nil
	}

	user, err := s.verifyOIDCToken(ctx.UserContext(), tokens.AccessToken)
	if err != nil {
		return false, fmt.Errorf("failed to verify token: %w", err)
	}

	if user == nil {
		return false, nil
	}

	ctx.Locals(UserKey.Value(), *user)

	expiresSoon := tokens.ExpiresIn.Add(-30 * time.Second).Before(time.Now().UTC())
	if !expiresSoon {
		return true, nil
	}

	if isRefreshing, ok := s.cookiesRefresh.Get(user.ID); isRefreshing && ok {
		return true, nil
	}

	s.cookiesRefresh.Set(user.ID, true)
	defer s.cookiesRefresh.Delete(user.ID)

	newTokens, err := s.refreshOIDCToken(ctx.UserContext(), tokens.RefreshToken)
	if err != nil {
		s.Logout(ctx)
		return false, fmt.Errorf("failed to refresh OIDC tokens: %w", err)
	}

	if err = s.setCookies(ctx, newTokens); err != nil {
		return false, fmt.Errorf("failed to set cookies: %w", err)
	}

	s.log.Debug().Str("user", user.Name).Msg("refreshed tokens in background")
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
		Name:     stateCookie,
		Value:    state,
		MaxAge:   stateCookieMaxAge,
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
	storedState := ctx.Cookies(stateCookie)
	if state == "" || state != storedState {
		return errInvalidState
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     stateCookie,
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

	rawIDToken, ok := token.Extra(idToken).(string)
	if !ok {
		return errors.New("missing id_token")
	}

	tokens := &OIDCTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      rawIDToken,
		ExpiresIn:    token.Expiry,
	}

	_, err = s.verifyOIDCToken(ctx.UserContext(), rawIDToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	if err = s.setCookies(ctx, tokens); err != nil {
		return err
	}

	return nil
}

func (s *cookieAuthService) parseLocalJWT(tokenString string) (*models.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &mpClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*mpClaims)
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
		return nil, errEmailNotVerified
	}

	user, err = s.users.GetByEmail(claims.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errCouldNotLinkUser
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
		ExpiresIn:    newToken.Expiry,
	}, nil
}

func (s *cookieAuthService) setCookies(ctx *fiber.Ctx, tokens *OIDCTokens) error {
	token, err := s.storeOidcToken(ctx.UserContext(), tokens)
	if err != nil {
		return err
	}

	ctx.Cookie(&fiber.Cookie{
		Name:     tokenCookie,
		Value:    token,
		MaxAge:   30 * 24 * 60 * 60, // 30 Days
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return nil
}

func (s *cookieAuthService) getOidcTokens(token string) (*OIDCTokens, error) {
	data, err := s.storage.Get(token)
	if err != nil {
		return nil, err
	}

	var oidcToken OIDCTokens
	if err = json.Unmarshal(data, &oidcToken); err != nil {
		return nil, err
	}

	return &oidcToken, nil
}

func (s *cookieAuthService) storeOidcToken(ctx context.Context, tokens *OIDCTokens) (string, error) {
	token, err := utils.GenerateSecret(32)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(&tokens)
	if err != nil {
		return "", err
	}

	if err = s.storage.Set(token, data, 30*24*time.Hour); err != nil {
		return "", err
	}

	return token, nil
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
	isAuthenticated, err := a.isAuthenticated(ctx)
	if err != nil {
		a.log.Warn().Err(err).Msg("error while checking api key auth")
	}
	if !isAuthenticated {
		return a.cookieAuth.Middleware(ctx)
	}
	return ctx.Next()
}

func (a *apiKeyAuthService) isAuthenticated(ctx *fiber.Ctx) (bool, error) {
	apiKey := ctx.Query(apiQueryKey)
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
