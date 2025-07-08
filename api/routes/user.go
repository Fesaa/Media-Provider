package routes

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRoutes struct {
	dig.In

	Router fiber.Router
	Auth   services.AuthService `name:"jwt-auth"`
	DB     *db.Database
	Log    zerolog.Logger

	Val             services.ValidationService
	Notify          services.NotificationService
	Transloco       services.TranslocoService
	SettingsService services.SettingsService
}

func RegisterUserRoutes(ur userRoutes) {

	ur.Router.Post("/login", ur.LoginUser)
	ur.Router.Post("/register", ur.RegisterUser)
	ur.Router.Get("/any-user-exists", ur.AnyUserExists)
	ur.Router.Post("/reset-password", ur.ResetPassword)

	user := ur.Router.Group("/user", ur.Auth.Middleware)
	user.Get("/refresh-api-key", ur.RefreshAPIKey)
	user.Get("/all", ur.Users)
	user.Post("/update", ur.UpdateUser)
	user.Delete("/:userId", ur.DeleteUser)
	user.Post("/reset/:userId", ur.GenerateResetPassword)
	user.Get("/me", ur.Me)
	user.Post("/me", ur.UpdateMe)
	user.Post("/password", ur.UpdatePassword)
}

func (ur *userRoutes) UpdatePassword(ctx *fiber.Ctx) error {
	var updatePasswordRequest payload.UpdatePasswordRequest
	if err := ur.Val.ValidateCtx(ctx, &updatePasswordRequest); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse update password request")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	user := ctx.Locals("user").(models.User)

	decodeString, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		ur.Log.Error().Err(err).Str("user", user.Name).Msg("failed to decode password")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	if err = bcrypt.CompareHashAndPassword(decodeString, []byte(updatePasswordRequest.OldPassword)); err != nil {
		ur.Log.Error().Err(err).Str("user", user.Name).Msg("invalid password")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid password",
		})
	}

	password, err := bcrypt.GenerateFromPassword([]byte(updatePasswordRequest.NewPassword), bcrypt.MinCost)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate password")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	user.PasswordHash = base64.StdEncoding.EncodeToString(password)

	if _, err = ur.DB.Users.Update(user); err != nil {
		ur.Log.Error().Err(err).Str("user", user.Name).Msg("failed to update user")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) UpdateMe(ctx *fiber.Ctx) error {
	var updateUserReq payload.UpdateUserRequest
	if err := ur.Val.ValidateCtx(ctx, &updateUserReq); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse update request")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	user := ctx.Locals("user").(models.User)

	if user.Name != updateUserReq.Name {
		other, err := ur.DB.Users.GetByName(updateUserReq.Name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		if other != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "user already exists",
			})
		}
	}

	user.Name = updateUserReq.Name

	if user.Email.String != updateUserReq.Email {
		other, err := ur.DB.Users.GetByEmail(updateUserReq.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		if other != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "user already exists",
			})
		}
	}

	user.Email = sql.NullString{String: updateUserReq.Email, Valid: true}

	if _, err := ur.DB.Users.Update(user); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) Me(ctx *fiber.Ctx) error {
	user, ok := ctx.Locals("user").(models.User)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "could not find user",
		})
	}

	return ctx.JSON(payload.LoginResponse{
		Id:          user.ID,
		Name:        user.Name,
		Email:       user.Email.String,
		ApiKey:      user.ApiKey,
		Permissions: user.Permission,
	})
}

func (ur *userRoutes) AnyUserExists(ctx *fiber.Ctx) error {
	ok, err := ur.DB.Users.ExistsAny()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to check if user exists")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	if ok {
		return ctx.SendString("true")
	}

	return ctx.SendString("false")
}

// Until we add a user service
//
//nolint:funlen
func (ur *userRoutes) RegisterUser(ctx *fiber.Ctx) error {
	ok, err := ur.DB.Users.ExistsAny()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to check if user exists")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	if ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}

	var register payload.LoginRequest
	if err := ur.Val.ValidateCtx(ctx, &register); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}

	password, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.MinCost)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate password")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate api key")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
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
				ur.Log.Warn().Err(err).Msg("failed to check existence of user, not setting all perms")
				return u
			}
			if ok {
				return u
			}

			u.Permission = models.ALL_PERMS
			u.Original = true
			return u
		})

	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to register user")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	loginRequest := payload.LoginRequest{
		UserName: user.Name,
		Password: register.Password,
		Remember: register.Remember,
	}

	res, err := ur.Auth.Login(loginRequest)
	if err != nil {
		return err
	}

	return ctx.JSON(res)
}

func (ur *userRoutes) LoginUser(ctx *fiber.Ctx) error {
	var login payload.LoginRequest
	if err := ur.Val.ValidateCtx(ctx, &login); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse body")
		return fiber.ErrBadRequest
	}

	settings, err := ur.SettingsService.GetSettingsDto()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to get settings")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get settings",
			"error":   err.Error(),
		})
	}

	if settings.Oidc.DisablePasswordLogin {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{})
	}

	res, err := ur.Auth.Login(login)
	if err != nil {
		return err
	}

	return ctx.JSON(res)
}

func (ur *userRoutes) RefreshAPIKey(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)

	key, err := utils.GenerateApiKey()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate api key")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	_, err = ur.DB.Users.Update(user, func(u models.User) models.User {
		u.ApiKey = key
		return u
	})

	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to refresh api key")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"ApiKey": key,
	})
}

func (ur *userRoutes) Users(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteUser) {
		return fiber.ErrForbidden
	}
	users, err := ur.DB.Users.All()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return ctx.JSON(utils.Map(users, func(u models.User) payload.UserDto {
		return payload.UserDto{
			ID:         u.ID,
			Name:       u.Name,
			Permission: u.Permission,
			CanDelete:  !u.Original,
		}
	}))
}

func (ur *userRoutes) UpdateUser(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteUser) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{})
	}

	var userDto payload.UserDto
	if err := ur.Val.ValidateCtx(ctx, &userDto); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse body")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var err error
	var newUser *models.User
	if userDto.ID != 0 {
		newUser, err = ur.DB.Users.UpdateById(userDto.ID, func(u models.User) models.User {
			u.Name = userDto.Name
			u.Permission = userDto.Permission
			return u
		})
	} else {
		newUser, err = ur.DB.Users.Create(userDto.Name, func(u models.User) models.User {
			u.Permission = userDto.Permission
			return u
		})
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	if newUser == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	return ctx.Status(fiber.StatusOK).JSON(payload.UserDto{
		ID:         newUser.ID,
		Name:       newUser.Name,
		Permission: newUser.Permission,
		CanDelete:  !newUser.Original,
	})
}

func (ur *userRoutes) DeleteUser(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermDeleteUser) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{})
	}

	userID, err := ParamsUInt(ctx, "userId")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	toDelete, err := ur.DB.Users.GetById(userID)
	if err != nil {
		ur.Log.Error().Uint("id", userID).Err(err).Msg("failed to check if user exists")
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": ur.Transloco.GetTranslation("user-not-found", userID),
		})
	}

	if toDelete.Original {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": ur.Transloco.GetTranslation("cant-delete-first-user"),
		})
	}

	err = ur.DB.Users.Delete(toDelete.ID)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to delete user")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) GenerateResetPassword(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteUser) {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{})
	}

	userId, err := ParamsUInt(ctx, "userId")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	resetUser, err := ur.DB.Users.GetById(userId)
	if err != nil {
		ur.Log.Error().Uint("id", userId).Err(err).Msg("failed to check if user exists")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if resetUser == nil {
		ur.Log.Error().Str("user", user.Name).Uint("id", userId).Err(err).Msg("user does not exist")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": ur.Transloco.GetTranslation("user-doesnt-exist"),
		})
	}

	reset, err := ur.DB.Users.GetResetByUserId(userId)
	if err == nil && reset != nil {
		fmt.Printf("A reset link has been found for %d, with key \"%s\"\nYou can surf to <.../login/reset?key=%s> to reset it.\n", reset.UserId, reset.Key, reset.Key)
		return ctx.Status(fiber.StatusOK).JSON(reset)
	} else if err != nil {
		ur.Log.Warn().Err(err).Msg("failed to check for existing reset key")
	}

	reset, err = ur.DB.Users.GenerateReset(userId)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate reset password")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	fmt.Printf("A reset link has been generated for %d, with key \"%s\"\nYou can surf to <.../login/reset?key=%s> to reset it.\n", reset.UserId, reset.Key, reset.Key)
	ur.Notify.NotifySecurityQ(
		ur.Transloco.GetTranslation("generate-reset-link-title"),
		ur.Transloco.GetTranslation("generate-reset-link-summary", user.Name, resetUser.Name))
	return ctx.Status(fiber.StatusOK).JSON(reset)
}

func (ur *userRoutes) ResetPassword(ctx *fiber.Ctx) error {
	var pl payload.ResetPasswordRequest
	if err := ur.Val.ValidateCtx(ctx, &pl); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse body")
		return fiber.ErrBadRequest
	}

	reset, err := ur.DB.Users.GetReset(pl.Key)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to check if user exists")
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": ur.Transloco.GetTranslation("failed-find-reset-key"),
		})
	}

	if reset == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": ur.Transloco.GetTranslation("failed-find-reset-key"),
		})
	}

	user, err := ur.DB.Users.GetById(reset.UserId)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": ur.Transloco.GetTranslation("failed-find-user"),
		})
	}
	if user == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	password, err := bcrypt.GenerateFromPassword([]byte(pl.Password), bcrypt.MinCost)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate password")
		return fiber.ErrInternalServerError
	}

	_, err = ur.DB.Users.Update(*user, func(u models.User) models.User {
		u.PasswordHash = base64.StdEncoding.EncodeToString(password)
		return u
	})

	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to update user password")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	if err = ur.DB.Users.DeleteReset(pl.Key); err != nil {
		ur.Log.Warn().Err(err).Msg("failed to delete reset key")
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
