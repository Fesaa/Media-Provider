package routes

import (
	"encoding/base64"
	"fmt"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type userRoutes struct {
	db *db.Database
}

func RegisterUserRoutes(router fiber.Router, db *db.Database, cache fiber.Handler) {
	ur := userRoutes{db: db}

	router.Post("/login", wrap(ur.LoginUser))
	router.Post("/register", wrap(ur.RegisterUser))
	router.Get("/any-user-exists", wrap(ur.AnyUserExists))
	router.Post("/reset-password", wrap(ur.ResetPassword))

	user := router.Group("/user", auth.Middleware)
	user.Get("/refresh-models-key", wrap(ur.RefreshApiKey))
	user.Get("/all", wrap(ur.Users))
	user.Post("/update", wrap(ur.UpdateUser))
	user.Delete("/:userId", wrap(ur.DeleteUser))
	user.Post("/reset/:userId", wrap(ur.GenerateResetPassword))
}

func (ur *userRoutes) AnyUserExists(l *log.Logger, ctx *fiber.Ctx) error {
	ok, err := ur.db.Users.ExistsAny()
	if err != nil {
		l.Error("failed to check existence of user", "err", err)
		return fiber.ErrInternalServerError
	}

	if ok {
		return ctx.SendString("true")
	}

	return ctx.SendString("false")
}

func (ur *userRoutes) RegisterUser(l *log.Logger, ctx *fiber.Ctx) error {
	var register payload.LoginRequest
	if err := ctx.BodyParser(&register); err != nil {
		l.Error("failed to parse body", "err", err)
		return fiber.ErrBadRequest
	}

	password, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.MinCost)
	if err != nil {
		l.Error("failed to hash password", "err", err)
		return fiber.ErrInternalServerError
	}

	apiKey, err := utils.GenerateApiKey()
	if err != nil {
		l.Error("failed to generate models key", "err", err)
		return fiber.ErrInternalServerError
	}

	user, err := ur.db.Users.Create(register.UserName,
		func(u models.User) models.User {
			u.PasswordHash = base64.StdEncoding.EncodeToString(password)
			u.ApiKey = apiKey
			return u
		},
		func(u models.User) models.User {
			var ok bool
			ok, err = ur.db.Users.ExistsAny()
			if err != nil {
				l.Warn("failed to check existence of user, not setting all perms", "err", err)
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
		l.Error("failed to create user", "err", err)
		return fiber.ErrInternalServerError
	}

	loginRequest := payload.LoginRequest{
		UserName: user.Name,
		Password: register.Password,
		Remember: register.Remember,
	}

	res, err := auth.I().Login(loginRequest)
	if err != nil {
		return err
	}

	return ctx.JSON(res)
}

func (ur *userRoutes) LoginUser(l *log.Logger, ctx *fiber.Ctx) error {
	var login payload.LoginRequest
	if err := ctx.BodyParser(&login); err != nil {
		l.Error("failed to parse body", "err", err)
		return fiber.ErrBadRequest
	}

	res, err := auth.I().Login(login)
	if err != nil {
		return err
	}

	return ctx.JSON(res)
}

func (ur *userRoutes) RefreshApiKey(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)

	key, err := utils.GenerateApiKey()
	if err != nil {
		l.Error("failed to generate models key", "err", err)
		return fiber.ErrInternalServerError
	}

	_, err = ur.db.Users.Update(user, func(u models.User) models.User {
		u.ApiKey = key
		return u
	})

	if err != nil {
		l.Error("failed to update user", "err", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendString(key)
}

func (ur *userRoutes) Users(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteUser) {
		return fiber.ErrForbidden
	}
	users, err := ur.db.Users.All()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(utils.Map(users, func(u models.User) payload.UserDto {
		return payload.UserDto{
			ID:         u.ID,
			Name:       u.Name,
			Permission: u.Permission,
		}
	}))
}

func (ur *userRoutes) UpdateUser(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteUser) {
		return fiber.ErrForbidden
	}

	var userDto payload.UserDto
	if err := ctx.BodyParser(&userDto); err != nil {
		l.Error("failed to parse body", "err", err)
		return fiber.ErrBadRequest
	}

	var err error
	var newUser *models.User
	if userDto.ID != 0 {
		newUser, err = ur.db.Users.UpdateById(userDto.ID, func(u models.User) models.User {
			u.Name = userDto.Name
			u.Permission = userDto.Permission
			return u
		})
	} else {
		newUser, err = ur.db.Users.Create(userDto.Name, func(u models.User) models.User {
			u.Permission = userDto.Permission
			return u
		})
	}

	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if newUser == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	return ctx.Status(fiber.StatusOK).JSON(newUser.ID)
}

func (ur *userRoutes) DeleteUser(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermDeleteUser) {
		return fiber.ErrForbidden
	}

	userId, _ := ctx.ParamsInt("userId", -1)
	if userId == -1 {
		return fiber.ErrBadRequest
	}

	toDelete, err := ur.db.Users.GetById(uint(userId))
	if err != nil {
		l.Error("could not find user specified in delete request", slog.Int("id", userId), "err", err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("user %d not found", userId),
		})
	}

	if toDelete.Original {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You may not delete the main user",
		})
	}

	err = ur.db.Users.Delete(toDelete.ID)
	if err != nil {
		l.Error("failed to delete user", "err", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) GenerateResetPassword(l *log.Logger, ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.User)
	if !user.HasPermission(models.PermWriteUser) {
		return fiber.ErrForbidden
	}

	userId, _ := ctx.ParamsInt("userId", -1)
	if userId == -1 {
		return fiber.ErrBadRequest
	}

	reset, err := ur.db.Users.GenerateReset(uint(userId))
	if err != nil {
		l.Error("failed to generate reset password", "err", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	fmt.Printf("A reset link has been generated for %d, with key \"%s\"\nYou can surf to <.../login/reset?key=%s> to reset it.", reset.UserId, reset.Key, reset.Key)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}

func (ur *userRoutes) ResetPassword(l *log.Logger, ctx *fiber.Ctx) error {
	var pl payload.ResetPasswordRequest
	if err := ctx.BodyParser(&pl); err != nil {
		l.Error("failed to parse body", "err", err)
		return fiber.ErrBadRequest
	}

	reset, err := ur.db.Users.GetReset(pl.Key)
	if err != nil {
		l.Error("an error occurred searching for the reset", "err", err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Failed to find reset key",
		})
	}

	if reset == nil {
		l.Warn("No reset found", "key", pl.Key)
		return fiber.ErrBadRequest
	}

	user, err := ur.db.Users.GetById(reset.UserId)
	if err != nil {
		l.Error("an error occurred searching for the user", "err", err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Failed to find user",
		})
	}
	if user == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{})
	}

	password, err := bcrypt.GenerateFromPassword([]byte(pl.Password), bcrypt.MinCost)
	if err != nil {
		l.Error("failed to hash password", "err", err)
		return fiber.ErrInternalServerError
	}

	_, err = ur.db.Users.Update(*user, func(u models.User) models.User {
		u.PasswordHash = base64.StdEncoding.EncodeToString(password)
		return u
	})

	if err != nil {
		l.Error("failed to update user", "err", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
	}

	if err = ur.db.Users.DeleteReset(pl.Key); err != nil {
		l.Warn("failed to delete reset key", "err", err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{})
}
