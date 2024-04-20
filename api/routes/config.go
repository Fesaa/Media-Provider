package routes

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/config"
	"github.com/gofiber/fiber/v2"
)

func ReloadPages(ctx *fiber.Ctx) error {
	err := config.ReloadPages("config.yaml")
	if err != nil {
		slog.Error("Failed to reload pages", "err", err)
		return fiber.ErrInternalServerError
	}
	return nil
}

func GetConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(config.C)
}
