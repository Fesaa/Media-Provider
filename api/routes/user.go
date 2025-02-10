package routes

import (
	"encoding/base64"
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
	"golang.org/x/crypto/bcrypt"
)

type userRoutes struct {
	dig.In

	Router fiber.Router
	Auth   auth.Provider `name:"jwt-auth"`
	DB     *db.Database
	Log    zerolog.Logger

	Val    services.ValidationService
	Notify services.NotificationService
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
}

func (ur *userRoutes) AnyUserExists(ctx *fiber.Ctx) error {
	ok, err := ur.DB.Users.ExistsAny()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to check if user exists")
		return fiber.ErrInternalServerError
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
		return fiber.ErrInternalServerError
	}

	if !ok {
		return fiber.ErrBadRequest
	}

	var register payload.LoginRequest
	if err := ur.Val.ValidateCtx(ctx, &register); err != nil {
		ur.Log.Error().Err(err).Msg("failed to parse body")
		return fiber.ErrBadRequest
	}

	password, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.MinCost)
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate password")
		return fiber.ErrInternalServerError
	}

	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		ur.Log.Error().Err(err).Msg("failed to generate api key")
		return fiber.ErrInternalServerError
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
		return fiber.ErrInternalServerError
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
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": fiber.ErrForbidden.Error(),
		})
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
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": fiber.ErrForbidden.Error(),
		})
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
			"message": fmt.Sprintf("user %d not found", userID),
		})
	}

	if toDelete.Original {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You may not delete the main user",
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
			"message": "User does not exist",
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
	ur.Notify.NotifySecurityQ("Reset link generated", fmt.Sprintf("%s has generated a reset link for %s", user.Name, resetUser.Name))
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
			"message": "Failed to find reset key",
		})
	}

	if reset == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Failed to find reset key",
		})
	}

	user, err := ur.DB.Users.GetById(reset.UserId)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Failed to find user",
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
