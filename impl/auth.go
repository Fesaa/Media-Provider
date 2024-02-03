package impl

import (
	"errors"
	"log/slog"
	"time"

	"github.com/Fesaa/Media-Provider/cerrors"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

const (
	TokenCookieName = "token"
)

type AuthImpl struct {
	authUser *models.User
}

func newAuth() *AuthImpl {
	return &AuthImpl{}
}

func (v *AuthImpl) IsAuthenticated(ctx *fiber.Ctx) (bool, error) {
	token := ctx.Cookies(TokenCookieName)
	if token == "" {
		return false, nil
	}

	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("No Holder found while handling auth. Was it set before AuthHandler was registered?")
		return false, errors.New("Internal Server Error.\nHolder was not present. Please contact the administrator.")
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		slog.Error("No DatabaseProvider found while handling auth. Was it implemented in the holderImpl?")
		return false, errors.New("Internal Server Error. \nNo DatabaseProvider found. Please contact the administrator.")
	}

	user, err := databaseProvider.GetUser(token)
	if err != nil {
		slog.Error("Error while getting user from database %s", err)
		return false, err
	}

	if user == nil {
		return false, nil
	}

	v.authUser = user

	return true, nil
}

type LoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember string `json:"remember,omitempty"`
}

func (v *AuthImpl) Login(ctx *fiber.Ctx) error {
	body := LoginBody{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return err
	}

	username := body.Username
	if username == "" {
		return badRequest("Username is required")
	}
	password := body.Password
	if password == "" {
		return badRequest("Password is required")
	}
	sessionOnly := body.Remember == ""

	token, err := v.login(ctx, username, password)
	if err != nil {
		return err
	}

	ctx.Cookie(&fiber.Cookie{
		Name:        TokenCookieName,
		Value:       *token,
		SessionOnly: sessionOnly,
	})
	return nil
}

func (v *AuthImpl) Logout(ctx *fiber.Ctx) error {
	ctx.Cookie(&fiber.Cookie{
		Name:    TokenCookieName,
		Expires: time.Now().Add(-(time.Hour * 5)),
	})
	v.authUser = nil
	return nil
}

func (v *AuthImpl) Register(ctx *fiber.Ctx) (*models.User, error) {
	body := LoginBody{}
	err := ctx.BodyParser(&body)
	if err != nil {
		return nil, err
	}

	username := body.Username
	if username == "" {
		return nil, badRequest("Username is required")
	}
	password := body.Password
	if password == "" {
		return nil, badRequest("Password is required")
	}
	sessionOnly := body.Remember == ""

	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		slog.Error("No Holder found while handling auth register. Was it set before AuthHandler was registered?")
		return nil, errors.New("No Holder found, cannot register. Contact an administrator.")
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		slog.Error("No DatabaseProvider found while handling auth register. Was it implemented in the holderImpl?")
		return nil, errors.New("No DatabaseProvider found, cannot register. Contact an administrator.")
	}

	user, token, err := databaseProvider.CreateUser(username, password)
	if err != nil {
		return nil, err
	}

	ctx.Cookie(&fiber.Cookie{
		Name:        TokenCookieName,
		Value:       *token,
		SessionOnly: sessionOnly,
	})
	return user, nil
}

func (v *AuthImpl) User(ctx *fiber.Ctx) *models.User {
	return v.authUser
}

func (v *AuthImpl) UserRaw(ctx *fiber.Ctx) (*models.User, error) {
	token := ctx.Cookies(TokenCookieName)
	if token == "" {
		return nil, nil
	}

	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		return nil, errors.New("No Holder found, cannot fetch user.")
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		return nil, errors.New("No DatabaseProvider found, cannot fetch user.")
	}

	user, err := databaseProvider.GetUser(token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (v AuthImpl) login(ctx *fiber.Ctx, username, password string) (*string, error) {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		return nil, errors.New("No Holder found, cannot login. Contact an administrator.")
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		return nil, errors.New("No DatabaseProvider found, cannot login. Contact an administrator.")
	}

	token, err := databaseProvider.GetToken(username, password)
	if err != nil {
		slog.Error("Error while getting token from database: " + err.Error())
		return nil, err
	}

	if token == nil {
		return nil, cerrors.InvalidCredentials
	}

	return token, nil
}

func badRequest(msg string) error {
	return &fiber.Error{
		Code:    fiber.ErrBadRequest.Code,
		Message: msg,
	}
}
