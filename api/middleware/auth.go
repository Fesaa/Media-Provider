package middleware

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
)

func HasRole(role models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals(services.UserKey).(models.User)
		if !ok {
			return fiber.ErrUnauthorized
		}

		if !user.Roles.HasRole(role) {
			return fiber.ErrForbidden
		}

		return c.Next()
	}
}
