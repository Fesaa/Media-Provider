package routes

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/dig"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	errUserAlreadyExists     = errors.New("user already exists")
	errPasswordLoginDisabled = errors.New("password login disabled")
	errInvalidCredentials    = errors.New("invalid username or password")
)

type userRoutes struct {
	dig.In

	Router     fiber.Router
	RootRouter fiber.Router `name:"root-router"`
	Auth       services.AuthService
	DB         *db.Database

	Val             services.ValidationService
	Notify          services.NotificationService
	Transloco       services.TranslocoService
	SettingsService services.SettingsService
}

func RegisterUserRoutes(ur userRoutes) {
	publicLimiter := limiter.New(limiter.Config{
		Max:                    10,
		Expiration:             time.Minute * 5,
		SkipSuccessfulRequests: true,
	})

	ur.RootRouter.
		Get("/oidc/login", ur.oidcLogin).
		Get("/oidc/callback", ur.oidcCallback)

	ur.Router.
		Get("/any-user-exists", ur.anyUserExists).
		Use(publicLimiter).
		Post("/login", withBodyValidation(ur.loginUser)).
		Post("/register", withBodyValidation(ur.registerUser)).
		Post("/reset-password", withBodyValidation(ur.resetPassword))

	user := ur.Router.Group("/user", ur.Auth.Middleware)

	user.
		Get("/refresh-api-key", ur.refreshAPIKey).
		Get("/me", ur.me).
		Post("/me", withBodyValidation(ur.updateMe)).
		Post("/password", withBodyValidation(ur.updatePassword))

	user.Use(hasRole(models.ManageUsers)).
		Get("/all", ur.users).
		Post("/update", withBodyValidation(ur.updateUser)).
		Delete("/:id", withParam(newIdPathParam(), ur.deleteUser)).
		Post("/reset/:id", withParam(newIdPathParam(), ur.generateResetPassword))
}

func (ur *userRoutes) updatePassword(ctx *fiber.Ctx, updatePasswordRequest payload.UpdatePasswordRequest) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := ctx.Locals("user").(models.User)

	decodeString, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		log.Error().Err(err).Str("user", user.Name).Msg("failed to decode password")
		return InternalError(err)
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(updatePasswordRequest.OldPassword)); err != nil {
		log.Error().Err(err).Str("user", user.Name).Msg("invalid password")
		return BadRequest()
	}

	password, err := bcrypt.GenerateFromPassword([]byte(updatePasswordRequest.NewPassword), bcrypt.MinCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate password")
		return InternalError(err)
	}

	user.PasswordHash = base64.StdEncoding.EncodeToString(password)

	if _, err = ur.DB.Users.Update(user); err != nil {
		log.Error().Err(err).Str("user", user.Name).Msg("failed to update user")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) updateMe(ctx *fiber.Ctx, updateUserReq payload.UpdateUserRequest) error {
	user := ctx.Locals("user").(models.User)

	if user.Name != updateUserReq.Name {
		other, err := ur.DB.Users.GetByName(updateUserReq.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return InternalError(err)
		}

		if other != nil {
			return BadRequest(errUserAlreadyExists)
		}
	}

	user.Name = updateUserReq.Name

	if user.Email.String != updateUserReq.Email {
		other, err := ur.DB.Users.GetByEmail(updateUserReq.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return InternalError(err)
		}

		if other != nil {
			return BadRequest(errUserAlreadyExists)
		}
	}

	user.Email = sql.NullString{String: updateUserReq.Email, Valid: true}

	if _, err := ur.DB.Users.Update(user); err != nil {
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) me(ctx *fiber.Ctx) error {
	user := services.GetFromContext(ctx, services.UserKey)

	return ctx.JSON(payload.LoginResponse{
		Id:     user.ID,
		Name:   user.Name,
		Email:  user.Email.String,
		ApiKey: user.ApiKey,
		Roles:  user.Roles,
	})
}

func (ur *userRoutes) anyUserExists(ctx *fiber.Ctx) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	ok, err := ur.DB.Users.ExistsAny()
	if err != nil {
		log.Error().Err(err).Msg("failed to check if user exists")
		return BadRequest() // Return bad request, unauthenticated endpoint
	}

	if ok {
		return ctx.SendString("true")
	}

	return ctx.SendString("false")
}

// Until we add a user service
//
//nolint:funlen
func (ur *userRoutes) registerUser(ctx *fiber.Ctx, register payload.LoginRequest) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	ok, err := ur.DB.Users.ExistsAny()
	if err != nil {
		log.Error().Err(err).Msg("failed to check if user exists")
		return BadRequest() // Return BadRequest as this is a public endpoint
	}

	if ok {
		return BadRequest()
	}

	password, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.MinCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate password")
		return InternalError(err)
	}

	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate api key")
		return InternalError(err)
	}

	user, err := ur.DB.Users.Create(register.UserName,
		func(u models.User) models.User {
			u.PasswordHash = base64.StdEncoding.EncodeToString(password)
			u.ApiKey = apiKey
			return u
		},
		func(u models.User) models.User {
			var ok bool
			ok, err = ur.DB.Users.ExistsAny()
			if err != nil {
				log.Warn().Err(err).Msg("failed to check existence of user, not setting all perms")
				return u
			}
			if ok {
				return u
			}

			u.Roles = models.AllRoles
			u.Original = true
			return u
		})

	if err != nil {
		log.Error().Err(err).Msg("failed to register user")
		return InternalError(err)
	}

	loginRequest := payload.LoginRequest{
		UserName: user.Name,
		Password: register.Password,
		Remember: register.Remember,
	}

	res, err := ur.Auth.Login(ctx, loginRequest)
	if err != nil {
		return InternalError(err)
	}

	return ctx.JSON(res)
}

func (ur *userRoutes) loginUser(ctx *fiber.Ctx, login payload.LoginRequest) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	settings, err := ur.SettingsService.GetSettingsDto()
	if err != nil {
		log.Error().Err(err).Msg("failed to get settings")
		return InternalError(err)
	}

	if settings.Oidc.DisablePasswordLogin {
		return BadRequest(errPasswordLoginDisabled)
	}

	res, err := ur.Auth.Login(ctx, login)
	if err != nil {
		log.Error().Err(err).Msg("failed to login")
		return BadRequest(errInvalidCredentials)
	}

	return ctx.JSON(res)
}

func (ur *userRoutes) logoutUser(ctx *fiber.Ctx) error {
	ur.Auth.Logout(ctx)
	return ctx.Redirect("/")
}

func (ur *userRoutes) oidcLogin(ctx *fiber.Ctx) error {
	url, err := ur.Auth.GetOIDCLoginURL(ctx)
	if err != nil {
		return InternalError(err)
	}

	return ctx.Redirect(url)
}

func (ur *userRoutes) oidcCallback(ctx *fiber.Ctx) error {
	if err := ur.Auth.HandleOIDCCallback(ctx); err != nil {
		return BadRequest(err)
	}

	redirectURL := ctx.Query("redirect")
	if redirectURL == "" {
		redirectURL = "/" // default redirect
	}
	return ctx.Redirect(redirectURL)
}

func (ur *userRoutes) refreshAPIKey(ctx *fiber.Ctx) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)

	key, err := utils.GenerateApiKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate api key")
		return InternalError(err)
	}

	_, err = ur.DB.Users.Update(user, func(u models.User) models.User {
		u.ApiKey = key
		return u
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to refresh api key")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"ApiKey": key,
	})
}

func (ur *userRoutes) users(ctx *fiber.Ctx) error {
	users, err := ur.DB.Users.All()
	if err != nil {
		return InternalError(err)
	}
	return ctx.JSON(utils.Map(users, func(u models.User) payload.UserDto {
		return payload.UserDto{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email.String,
			Roles:     u.Roles,
			Pages:     u.Pages,
			CanDelete: !u.Original,
		}
	}))
}

func (ur *userRoutes) updateUser(ctx *fiber.Ctx, userDto payload.UserDto) error {
	var err error
	var newUser *models.User
	if userDto.ID != 0 {
		newUser, err = ur.DB.Users.UpdateById(userDto.ID, func(u models.User) models.User {
			u.Name = userDto.Name
			u.Email = sql.NullString{String: userDto.Email, Valid: true}
			if !u.Original {
				u.Roles = userDto.Roles
				u.Pages = userDto.Pages
			}

			return u
		})
	} else {
		newUser, err = ur.DB.Users.Create(userDto.Name, func(u models.User) models.User {
			u.Roles = userDto.Roles
			return u
		})
	}

	if err != nil {
		return InternalError(err)
	}

	if newUser == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	return ctx.Status(fiber.StatusOK).JSON(payload.UserDto{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Roles:     newUser.Roles,
		Pages:     newUser.Pages,
		CanDelete: !newUser.Original,
	})
}

func (ur *userRoutes) deleteUser(ctx *fiber.Ctx, userID int) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	toDelete, err := ur.DB.Users.GetById(userID)
	if err != nil {
		log.Error().Int("id", userID).Err(err).Msg("failed to check if user exists")
		return NotFound(errors.New(ur.Transloco.GetTranslation("user-not-found", userID)))
	}

	if toDelete.Original {
		return BadRequest(errors.New(ur.Transloco.GetTranslation("cant-delete-first-user")))
	}

	err = ur.DB.Users.Delete(toDelete.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete user")
		return InternalError(err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) generateResetPassword(ctx *fiber.Ctx, userId int) error {
	log := services.GetFromContext(ctx, services.LoggerKey)
	user := services.GetFromContext(ctx, services.UserKey)

	resetUser, err := ur.DB.Users.GetById(userId)
	if err != nil {
		log.Error().Int("id", userId).Err(err).Msg("failed to check if user exists")
		return InternalError(err)
	}

	if resetUser == nil {
		log.Error().Str("user", user.Name).Int("id", userId).Err(err).Msg("user does not exist")
		return NotFound()
	}

	reset, err := ur.DB.Users.GetResetByUserId(userId)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for existing reset key")
	}

	if reset != nil {
		fmt.Printf("A reset link has been found for %d, with key \"%s\"\nYou can surf to <.../login/reset?key=%s> to reset it.\n", reset.UserId, reset.Key, reset.Key)
		return ctx.Status(fiber.StatusOK).JSON(reset)
	}

	reset, err = ur.DB.Users.GenerateReset(userId)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate reset password")
		return InternalError(err)
	}

	fmt.Printf("A reset link has been generated for %d, with key \"%s\"\nYou can surf to <.../login/reset?key=%s> to reset it.\n", reset.UserId, reset.Key, reset.Key)
	ur.Notify.Notify(models.NewNotification().
		WithTitle(ur.Transloco.GetTranslation("generate-reset-link-title")).
		WithBody(ur.Transloco.GetTranslation("generate-reset-link-summary", user.Name, resetUser.Name)).
		WithGroup(models.GroupSecurity).
		WithColour(models.Warning).
		WithRequiredRoles(models.ManageUsers).
		Build())
	return ctx.Status(fiber.StatusOK).JSON(reset)
}

func (ur *userRoutes) resetPassword(ctx *fiber.Ctx, pl payload.ResetPasswordRequest) error {
	log := services.GetFromContext(ctx, services.LoggerKey)

	reset, err := ur.DB.Users.GetReset(pl.Key)
	if err != nil {
		log.Error().Err(err).Str("key", pl.Key).Msg("failed to check if user exists")
		return NotFound()
	}

	if reset == nil {
		log.Warn().Str("key", pl.Key).Msg("Reset request with invalid key")
		return NotFound()
	}

	user, err := ur.DB.Users.GetById(reset.UserId)
	if err != nil {
		return NotFound()
	}
	if user == nil {
		return NotFound()
	}

	password, err := bcrypt.GenerateFromPassword([]byte(pl.Password), bcrypt.MinCost)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate password")
		return InternalError(err)
	}

	_, err = ur.DB.Users.Update(*user, func(u models.User) models.User {
		u.PasswordHash = base64.StdEncoding.EncodeToString(password)
		return u
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to update user password")
		return InternalError(err)
	}

	if err = ur.DB.Users.DeleteReset(pl.Key); err != nil {
		log.Warn().Err(err).Msg("failed to delete reset key")
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
