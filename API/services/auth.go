package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	fiberutils "github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

	// Login perform local Auth login
	Login(*fiber.Ctx, payload.LoginRequest) (*payload.LoginResponse, error)
	// Logout deletes the cookies, and return the logout url if applicable (oidc)
	Logout(*fiber.Ctx) string
	// GetOIDCLoginURL returns the redirect url of the oidc provider
	GetOIDCLoginURL(*fiber.Ctx) (string, error)
	// HandleOIDCCallback authenticates with oidc and sets the correct cookies
	HandleOIDCCallback(*fiber.Ctx) error
}

type OIDCTokens struct {
	UserId       int       `json:"user_id"`
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
	unitOfWork *db.UnitOfWork
	cfg        *config.Config
	log        zerolog.Logger
	storage    CacheService

	httpClient *http.Client

	oidcSettings payload.OidcSettings
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier

	cookiesRefresh utils.SafeMap[int, bool]
}

type cookieAuthServiceParams struct {
	dig.In

	Ctx        context.Context
	UnitOfWork *db.UnitOfWork
	Service    SettingsService
	Storage    CacheService
	Config     *config.Config
	Log        zerolog.Logger
	HttpClient *menou.Client
}

func CookieAuthServiceProvider(params cookieAuthServiceParams) (AuthService, error) {
	ctx := oidc.ClientContext(params.Ctx, params.HttpClient.Client)
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanSetupService,
		trace.WithAttributes(attribute.String("service.name", "CookieAuthService")))
	defer span.End()

	settings, err := params.Service.GetSettingsDto(ctx)
	if err != nil {
		return nil, err
	}

	s := &cookieAuthService{
		unitOfWork:     params.UnitOfWork,
		storage:        params.Storage,
		cfg:            params.Config,
		cookiesRefresh: utils.NewSafeMap[int, bool](),
		httpClient:     params.HttpClient.Client,
		log:            params.Log.With().Str("handler", "cookie-auth-service").Logger(),
	}

	if err = s.setupOIDC(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to setup OIDC: %w", err)
	}

	return s, nil
}

func (s *cookieAuthService) setupOIDC(ctx context.Context, settings payload.Settings) error {
	if !settings.Oidc.Enabled() {
		return nil
	}

	provider, err := oidc.NewProvider(ctx, settings.Oidc.Authority)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	s.oidcProvider = provider
	s.oidcSettings = settings.Oidc
	s.verifier = provider.Verifier(&oidc.Config{ClientID: settings.Oidc.ClientID})

	s.oauth2Config = &oauth2.Config{
		ClientID:     settings.Oidc.ClientID,
		ClientSecret: settings.Oidc.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", oidc.ScopeOfflineAccess},
	}

	s.log.Debug().
		Str("authority", settings.Oidc.Authority).
		Str("client_id", settings.Oidc.ClientID).
		Msg("OIDC configured successfully")
	return nil
}

func (s *cookieAuthService) Login(ctx *fiber.Ctx, loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	user, err := s.unitOfWork.Users.GetByName(ctx.UserContext(), loginRequest.UserName)
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

	if err = s.setCookies(ctx, &OIDCTokens{UserId: user.ID}); err != nil {
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

func (s *cookieAuthService) Logout(ctx *fiber.Ctx) string {
	ctx.Cookie(&fiber.Cookie{
		Name:     tokenCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	ctx.Cookie(&fiber.Cookie{
		Name:     stateCookie,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	token := ctx.Cookies(tokenCookie)
	if token == "" {
		return ""
	}

	logoutUrl := s.getSessionLogoutUrl(ctx, token) // Must happen before we delete the tokens
	if err := s.storage.DeleteWithContext(ctx.UserContext(), token); err != nil {
		s.log.Warn().Err(err).Msg("failed to delete token during logout")
	}

	return logoutUrl
}

func (s *cookieAuthService) getSessionLogoutUrl(ctx *fiber.Ctx, token string) string {
	if s.oidcProvider == nil {
		return ""
	}

	tokens, err := s.getOidcTokens(ctx.UserContext(), token)
	if err != nil {
		s.log.Error().Err(err).Str("token", token).Msg("failed to get token")
	}

	if tokens == nil || tokens.IDToken == "" {
		return ""
	}

	var claims struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}

	if err = s.oidcProvider.Claims(&claims); err != nil {
		s.log.Error().Err(err).Msg("Discovery document failed parsing")
		return ""
	}

	if claims.EndSessionEndpoint == "" {
		return ""
	}

	logoutParams := url.Values{}
	logoutParams.Add("id_token_hint", tokens.IDToken)
	postLogoutUrl := s.getUrlBase(ctx) + "/login"
	logoutParams.Add("post_logout_redirect_uri", postLogoutUrl)

	return claims.EndSessionEndpoint + "?" + logoutParams.Encode()
}

func (s *cookieAuthService) Middleware(ctx *fiber.Ctx) error {
	if !s.isAuthenticated(ctx) {
		s.Logout(ctx)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Next()
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
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   ctx.Secure(),
	})

	s.oauth2Config.RedirectURL = s.getUrlBase(ctx) + "/oidc/callback"
	loginUrl := s.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return loginUrl, nil
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

	user, err := s.verifyOIDCToken(ctx.UserContext(), rawIDToken)
	if err != nil {
		return fmt.Errorf("failed to verify ID token: %w", err)
	}

	tokens.UserId = user.ID
	if err = s.setCookies(ctx, tokens); err != nil {
		return err
	}

	return nil
}

func (s *cookieAuthService) deleteToken(ctx context.Context, authenticationFailed bool, token string) {
	if authenticationFailed {
		if err := s.storage.DeleteWithContext(ctx, token); err != nil {
			s.log.Warn().Err(err).Str("token", token).Msg("failed to delete token")
		}
	}
}

func (s *cookieAuthService) isAuthenticated(ctx *fiber.Ctx) (success bool) {
	token := ctx.Cookies(tokenCookie)
	if token == "" {
		return false
	}
	defer s.deleteToken(ctx.UserContext(), success, token)

	tokens, err := s.getOidcTokens(ctx.UserContext(), token)
	if err != nil {
		return false
	}

	user, err := s.unitOfWork.Users.GetByID(ctx.UserContext(), tokens.UserId)
	if user == nil || err != nil {
		return false
	}

	if tokens.AccessToken == "" { // Local auth only stores user id
		return true
	}

	if s.verifier == nil {
		return false
	}

	contextkey.SetInContext(ctx, contextkey.User, *user)

	expiresSoon := tokens.ExpiresIn.Add(-30 * time.Second).Before(time.Now().UTC())
	if !expiresSoon {
		return true
	}

	if isRefreshing, ok := s.cookiesRefresh.Get(user.ID); isRefreshing && ok {
		return true
	}

	err = s.refreshToken(ctx, token, tokens)
	if err != nil {
		s.log.Warn().Err(err).Msg("failed to refresh token")
	}
	return err == nil
}

func (s *cookieAuthService) refreshToken(fiberCtx *fiber.Ctx, oldKey string, tokens *OIDCTokens) error {
	ctx, span := tracing.TracerServices.Start(fiberCtx.UserContext(), tracing.SpanServicesOIDCTokenRefresh)
	defer span.End()

	s.cookiesRefresh.Set(tokens.UserId, true)
	defer s.cookiesRefresh.Delete(tokens.UserId)

	newTokens, err := s.refreshOIDCToken(ctx, tokens.UserId, tokens.RefreshToken)
	if err != nil {
		s.Logout(fiberCtx)
		return fmt.Errorf("failed to refresh OIDC tokens: %w", err)
	}

	if err = s.storage.DeleteWithContext(ctx, oldKey); err != nil {
		s.log.Warn().Err(err).Msg("failed to delete old token")
	}

	if err = s.setCookies(fiberCtx, newTokens); err != nil {
		return fmt.Errorf("failed to set cookies: %w", err)
	}

	s.log.Debug().Int("user", tokens.UserId).
		Time("expires_at", newTokens.ExpiresIn).
		Msg("refreshed tokens in background")
	return nil
}

func (s *cookieAuthService) verifyOIDCToken(ctx context.Context, tokenString string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	token, err := s.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.unitOfWork.Users.GetByExternalID(ctx, token.Subject)
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

	user, err = s.unitOfWork.Users.GetByEmail(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errCouldNotLinkUser
	}

	user.ExternalId = sql.NullString{String: token.Subject, Valid: true}
	if err = s.unitOfWork.Users.Update(ctx, *user); err != nil {
		s.log.Error().Err(err).
			Str("email", claims.Email).
			Msg("failed to assign external id to user")
		return nil, err
	}
	return user, nil
}

func (s *cookieAuthService) refreshOIDCToken(ctx context.Context, userId int, refreshToken string) (*OIDCTokens, error) {
	span := trace.SpanFromContext(ctx)

	ctx = oidc.ClientContext(ctx, s.httpClient)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if s.oauth2Config == nil {
		return nil, errors.New("OIDC not configured")
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := s.oauth2Config.TokenSource(ctx, token).Token()
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	rawIDToken, ok := newToken.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("missing id_token in refresh response")
	}

	return &OIDCTokens{
		UserId:       userId,
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
		SameSite: fiber.CookieSameSiteStrictMode,
		Secure:   ctx.Secure(),
	})

	return nil
}

func (s *cookieAuthService) getOidcTokens(ctx context.Context, token string) (*OIDCTokens, error) {
	data, err := s.storage.GetWithContext(ctx, token)
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

	if err = s.storage.SetWithContext(ctx, token, data, 30*24*time.Hour); err != nil {
		return "", err
	}

	return token, nil
}

func (s *cookieAuthService) getUrlBase(ctx *fiber.Ctx) string {
	scheme := "http"
	if !config.Development {
		scheme = "https"
	}
	return scheme + "://" + fiberutils.CopyString(ctx.Hostname())
}

func ApiKeyAuthServiceProvider(params apiKeyAuthServiceParams) AuthMiddleware {
	return &apiKeyAuthService{
		unitOfWork: params.UnitOfWork,
		cookieAuth: params.FallbackAuth,
		log:        params.Log.With().Str("handler", "api-key-auth-service").Logger(),
	}
}

type apiKeyAuthServiceParams struct {
	dig.In

	UnitOfWork   *db.UnitOfWork
	FallbackAuth AuthService
	Log          zerolog.Logger
}

type apiKeyAuthService struct {
	unitOfWork *db.UnitOfWork
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

	user, err := a.unitOfWork.Users.GetByAPIKey(ctx.UserContext(), apiKey)
	if err != nil {
		return false, err
	}

	contextkey.SetInContext(ctx, contextkey.User, *user)
	return true, nil
}
