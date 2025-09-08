package routes

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
)

func hasRole(role models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := contextkey.GetFromContext(c, contextkey.Logger)
		user, ok := contextkey.GetFromContextSafe(c, contextkey.User)
		if !ok {
			return fiber.ErrUnauthorized
		}

		if !user.Roles.HasRole(role) {
			log.Warn().Str("user", user.Name).Strs("roles", utils.MapToString(user.Roles)).
				Msg("user tried to access content without required roles")
			return fiber.ErrForbidden
		}

		return c.Next()
	}
}
