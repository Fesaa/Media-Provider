package routes

import (
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func GetPerms(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals("holder").(models.Holder)
	if !ok {
		slog.Error("Holder not found while refreshing permissions")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		slog.Error("Database provider not found while refreshing permissions")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	permissionProvider := databaseProvider.GetPermissionProvider()
	if permissionProvider == nil {
		slog.Error("Permission provider not found while refreshing permissions")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	perms := permissionProvider.GetAllPermissions()
	return ctx.JSON(perms)
}

func RefreshPerms(ctx *fiber.Ctx) error {
	holder, ok := ctx.Locals("holder").(models.Holder)
	if !ok {
		slog.Error("Holder not found while refreshing permissions")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		slog.Error("Database provider not found while refreshing permissions")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	permissionProvider := databaseProvider.GetPermissionProvider()
	if permissionProvider == nil {
		slog.Error("Permission provider not found while refreshing permissions")
		return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	err := permissionProvider.RefreshPermissions()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).SendString("Unable to refresh permissions: " + err.Error())
	}

	return ctx.Status(fiber.StatusOK).SendString("Permissions refreshed")
}
