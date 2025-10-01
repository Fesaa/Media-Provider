package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/http/menou"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/internal/tracing"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
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
	// HandleOIDCCallback authenticates with oidc and sets the correct cookies
	HandleOIDCCallback(*fiber.Ctx) error
}

type cookieAuthService struct {
	unitOfWork  *db.UnitOfWork
	cfg         *config.Config
	log         zerolog.Logger
	storage     CacheService
	userService UserService

	httpClient *http.Client

	cookiesRefresh utils.SafeMap[int, bool]
}

type cookieAuthServiceParams struct {
	dig.In

	Ctx        context.Context
	UnitOfWork *db.UnitOfWork
	Service    SettingsService
	Storage    CacheService
	User       UserService
	Config     *config.Config
	Log        zerolog.Logger
	HttpClient *menou.Client
}

func CookieAuthServiceProvider(params cookieAuthServiceParams) AuthService {
	return &cookieAuthService{
		unitOfWork:     params.UnitOfWork,
		storage:        params.Storage,
		cfg:            params.Config,
		userService:    params.User,
		cookiesRefresh: utils.NewSafeMap[int, bool](),
		httpClient:     params.HttpClient.Client,
		log:            params.Log.With().Str("handler", "cookie-auth-service").Logger(),
	}
}

func (s *cookieAuthService) Login(ctx *fiber.Ctx, loginRequest payload.LoginRequest) (*payload.LoginResponse, error) {
	user, err := s.userService.CheckPassword(ctx.UserContext(), loginRequest.UserName, loginRequest.Password)
	if err != nil {
		return nil, err
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
	deleteCookie(ctx, tokenCookie)
	deleteCookie(ctx, stateCookie)

	token := ctx.Cookies(tokenCookie)
	if token == "" {
		return ""
	}

	tokens, err := s.getOidcTokens(ctx.UserContext(), token)
	if err != nil {
		return ""
	}

	logoutUrl := s.userService.OidcLogoutUrl(ctx, tokens)
	if err = s.storage.DeleteWithContext(ctx.UserContext(), token); err != nil {
		s.log.Warn().Err(err).Msg("failed to delete token during logout")
	}

	return logoutUrl
}

func (s *cookieAuthService) Middleware(ctx *fiber.Ctx) error {
	if !s.isAuthenticated(ctx) {
		s.Logout(ctx)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	return ctx.Next()
}

func (s *cookieAuthService) HandleOIDCCallback(ctx *fiber.Ctx) error {
	tokens, err := s.userService.OidcLogin(ctx)
	if err != nil {
		return err
	}

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
		contextkey.SetInContext(ctx, contextkey.User, *user)
		return true
	}

	if !s.userService.OidcEnabled() {
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

	newTokens, err := s.userService.OidcRefreshToken(ctx, tokens.UserId, tokens.RefreshToken)
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
