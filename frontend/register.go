package frontend

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func register(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
	if !ok {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	authProvider := holder.GetAuthProvider()
	if authProvider == nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	auth, err := authProvider.IsAuthenticated(ctx)
	if err != nil {
		slog.Error(err.Error())
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if auth {
		return ctx.Redirect("/", 200)
	}

	return ctx.Render("register", fiber.Map{})
}
