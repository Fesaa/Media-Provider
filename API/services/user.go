package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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
	"github.com/Fesaa/Media-Provider/internal/tracing"
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

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrOidcNotSupported   = errors.New("oidc not supported")
	ErrInvalidState       = errors.New("oidc: invalid state")
	ErrNoOidcCode         = errors.New("oidc: no code")
	ErrEmailNotVerified   = errors.New("oidc: email not verified")
)

type OIDCTokens struct {
	UserId       int       `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token"`
	ExpiresIn    time.Time `json:"expires_in"`
}

type UserService interface {
	CheckPassword(ctx context.Context, username string, password string) (*models.User, error)

	OidcEnabled() bool
	OidcLogin(ctx *fiber.Ctx) (*OIDCTokens, error)
	OidcLoginUrl(ctx *fiber.Ctx) (string, error)
	OidcLogoutUrl(ctx *fiber.Ctx, tokens *OIDCTokens) string
	OidcRefreshToken(ctx context.Context, userId int, refreshToken string) (*OIDCTokens, error)
}

type userService struct {
	log zerolog.Logger

	unitOfWork *db.UnitOfWork
	httpClient *http.Client

	oidcSettings payload.OidcSettings
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

type userServiceParams struct {
	dig.In

	Ctx        context.Context
	UnitOfWork *db.UnitOfWork
	Service    SettingsService
	Config     *config.Config
	Log        zerolog.Logger
	HttpClient *menou.Client
}

func UserServiceProvider(params userServiceParams) (UserService, error) {
	ctx := oidc.ClientContext(params.Ctx, params.HttpClient.Client)
	ctx, span := tracing.TracerServices.Start(ctx, tracing.SpanSetupService,
		trace.WithAttributes(attribute.String("service.name", "UserService")))
	defer span.End()

	settings, err := params.Service.GetSettingsDto(ctx)
	if err != nil {
		return nil, err
	}

	s := &userService{
		log:        params.Log.With().Str("handler", "user-service").Logger(),
		unitOfWork: params.UnitOfWork,
		httpClient: params.HttpClient.Client,
	}

	if err = s.setupOIDC(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to setup OIDC: %w", err)
	}

	return s, nil
}

func (u *userService) setupOIDC(ctx context.Context, settings payload.Settings) error {
	if !settings.Oidc.Enabled() {
		return nil
	}

	provider, err := oidc.NewProvider(ctx, settings.Oidc.Authority)
	if err != nil {
		return fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	u.oidcProvider = provider
	u.oidcSettings = settings.Oidc
	u.verifier = provider.Verifier(&oidc.Config{ClientID: settings.Oidc.ClientID})

	u.oauth2Config = &oauth2.Config{
		ClientID:     settings.Oidc.ClientID,
		ClientSecret: settings.Oidc.ClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", oidc.ScopeOfflineAccess},
	}

	u.log.Debug().
		Str("authority", settings.Oidc.Authority).
		Str("client_id", settings.Oidc.ClientID).
		Msg("OIDC configured successfully")
	return nil
}

func (u *userService) CheckPassword(ctx context.Context, username string, password string) (*models.User, error) {
	user, err := u.unitOfWork.Users.GetByName(ctx, username)
	if err != nil {
		u.log.Error().Err(err).Str("user", username).Msg("user not found")
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	decodeString, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (u *userService) OidcEnabled() bool {
	return u.verifier != nil
}

func (u *userService) OidcRefreshToken(ctx context.Context, userId int, refreshToken string) (*OIDCTokens, error) {
	span := trace.SpanFromContext(ctx)

	ctx = oidc.ClientContext(ctx, u.httpClient)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if !u.OidcEnabled() {
		return nil, ErrOidcNotSupported
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	newToken, err := u.oauth2Config.TokenSource(ctx, token).Token()
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

func (u *userService) OidcLogoutUrl(ctx *fiber.Ctx, tokens *OIDCTokens) string {
	if u.oidcProvider == nil {
		return ""
	}

	var claims struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}

	if err := u.oidcProvider.Claims(&claims); err != nil {
		u.log.Error().Err(err).Msg("Discovery document failed parsing")
		return ""
	}

	if claims.EndSessionEndpoint == "" {
		return ""
	}

	logoutParams := url.Values{}
	logoutParams.Add("id_token_hint", tokens.IDToken)
	postLogoutUrl := u.getUrlBase(ctx) + "/login"
	logoutParams.Add("post_logout_redirect_uri", postLogoutUrl)

	return claims.EndSessionEndpoint + "?" + logoutParams.Encode()
}

func (u *userService) OidcLoginUrl(ctx *fiber.Ctx) (string, error) {
	if u.oauth2Config == nil {
		return "", errors.New("OIDC not configured")
	}

	b := make([]byte, 32)
	_, _ = rand.Read(b) // Read does not return errors
	state := base64.URLEncoding.EncodeToString(b)

	ctx.Cookie(&fiber.Cookie{
		Name:     stateCookie,
		Value:    state,
		MaxAge:   stateCookieMaxAge,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   ctx.Secure(),
	})

	u.oauth2Config.RedirectURL = u.getUrlBase(ctx) + "/oidc/callback"
	loginUrl := u.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return loginUrl, nil
}

func (u *userService) OidcLogin(ctx *fiber.Ctx) (*OIDCTokens, error) {
	if u.oauth2Config == nil {
		return nil, ErrOidcNotSupported
	}

	state := ctx.Query("state")
	storedState := ctx.Cookies(stateCookie)
	if state == "" || state != storedState {
		return nil, ErrInvalidState
	}

	deleteCookie(ctx, stateCookie)

	code := ctx.Query("code")
	if code == "" {
		return nil, ErrNoOidcCode
	}

	token, err := u.oauth2Config.Exchange(ctx.UserContext(), code)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to exchange token: %w", err)
	}

	rawIDToken, ok := token.Extra(idToken).(string)
	if !ok {
		return nil, errors.New("missing id_token")
	}

	tokens := &OIDCTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		IDToken:      rawIDToken,
		ExpiresIn:    token.Expiry,
	}

	user, err := u.getOrCreateUser(ctx.UserContext(), rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to get user: %w", err)
	}

	tokens.UserId = user.ID
	return tokens, nil
}

func (u *userService) getOrCreateUser(ctx context.Context, tokenString string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	token, err := u.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	user, err := u.unitOfWork.Users.GetByExternalID(ctx, token.Subject)
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
		return nil, fmt.Errorf("oidc: failed to parse user claims: %w", err)
	}

	if !claims.Verified {
		return nil, ErrEmailNotVerified
	}

	user, err = u.unitOfWork.Users.GetByEmail(ctx, claims.Email)
	if err != nil {
		return nil, err
	}

	if user != nil {
		user.ExternalId = sql.NullString{String: token.Subject, Valid: true}
		if err = u.unitOfWork.Users.Update(ctx, *user); err != nil {
			u.log.Error().Err(err).
				Str("email", claims.Email).
				Msg("failed to assign external id to user")
			return nil, fmt.Errorf("oidc: failed to assign external id to user: %w", err)
		}
	}

	// TODO: Create user based on server setting

	return nil, ErrUserNotFound
}

func (u *userService) getUrlBase(ctx *fiber.Ctx) string {
	scheme := "http"
	if !config.Development {
		scheme = "https"
	}
	return scheme + "://" + fiberutils.CopyString(ctx.Hostname())
}

func deleteCookie(ctx *fiber.Ctx, name string) {
	ctx.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
}
