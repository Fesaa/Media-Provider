package routes

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

// TODO: Switch over to our own context once we get our .NET approach finished

type Handler func(Ctx) error

type Ctx struct {
	*fiber.Ctx

	user *models.User
	log  zerolog.Logger
}

func InterOp(handler Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := contextkey.GetFromContext(c, contextkey.Logger)
		user, _ := contextkey.GetFromContextSafe(c, contextkey.User)

		ctx := Ctx{
			Ctx:  c,
			user: &user,
			log:  log,
		}

		return handler(ctx)
	}
}
