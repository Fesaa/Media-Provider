package routes

import (
	"encoding/base64"
	"github.com/Fesaa/Media-Provider/auth"
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/Fesaa/Media-Provider/payload"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func AnyUserExists(ctx *fiber.Ctx) error {
	ok, err := models.AnyUserExists()
	if err != nil {
		log.Error("failed to check existence of user", "err", err)
		return fiber.ErrInternalServerError
	}

	if ok {
		return ctx.SendStatus(fiber.StatusOK)
	}

	return ctx.SendStatus(fiber.StatusNotFound)
}

func RegisterUser(l *log.Logger, ctx *fiber.Ctx) error {
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
		l.Error("failed to generate api key", "err", err)
		return fiber.ErrInternalServerError
	}

	user, err := models.CreateUser(register.UserName,
		func(u *models.User) *models.User {
			u.PasswordHash = base64.StdEncoding.EncodeToString(password)
			u.ApiKey = apiKey
			return u
		})

	if err != nil {
		l.Error("failed to create user", "err", err)
		return fiber.ErrInternalServerError
	}

	loginRequest := payload.LoginRequest{
		UserName: user.Name,
		Password: user.PasswordHash,
		Remember: register.Remember,
	}

	res, err := auth.I().Login(loginRequest)
	if err != nil {
		return err
	}

	return ctx.JSON(res)
}

func LoginUser(l *log.Logger, ctx *fiber.Ctx) error {
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

func RefreshApiKey(l *log.Logger, ctx *fiber.Ctx) error {
	userName, _ := ctx.Locals("user").(string)

	key, err := utils.GenerateApiKey()
	if err != nil {
		l.Error("failed to generate api key", "err", err)
		return fiber.ErrInternalServerError
	}

	user, err := models.GetUser(userName)
	if err != nil {
		l.Error("failed to get user", "err", err)
		return fiber.ErrInternalServerError
	}

	_, err = models.UpdateUser(user, func(u *models.User) *models.User {
		u.ApiKey = key
		return u
	})

	if err != nil {
		l.Error("failed to update user", "err", err)
		return fiber.ErrInternalServerError
	}

	return ctx.SendString(key)
}
