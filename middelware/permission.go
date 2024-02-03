package middleware

import (
	"fmt"
	"log/slog"

	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func HasPermissions(holder models.Holder, permissions ...string) func(ctx *fiber.Ctx) error {
	if len(permissions) == 0 {
		return func(ctx *fiber.Ctx) error {
			return ctx.Next()
		}
	}

	databaseProvider := holder.GetDatabaseProvider()
	if databaseProvider == nil {
		panic("No DatabaseProvider found while handling permissions. Was it implemented in the holderImpl?")
	}

	permissionProvider := databaseProvider.GetPermissionProvider()
	if permissionProvider == nil {
		panic("No PermissionProvider found while handling permissions. Was it implemented in the holderImpl?")
	}

	var perms []*models.Permission = make([]*models.Permission, len(permissions))
	for i, perm := range permissions {
		p := permissionProvider.GetPermissionByKey(perm)
		if p == nil {
			slog.Warn(fmt.Sprintf("Permission with key %s was not found. Will be assumed to be missing for everyone.", perm))
		}
		perms[i] = p
	}

	return func(ctx *fiber.Ctx) error {
		holder, ok := ctx.Locals(models.HolderKey).(models.Holder)
		if !ok {
			slog.Error("No Holder found while handling permissions. Was it set before HasPermissions was registered?")
			return ctx.Status(500).SendString("Internal Server Error.\nHolder was not present. Please contact the administrator.")
		}

		authProvider := holder.GetAuthProvider()
		if authProvider == nil {
			slog.Error("No AuthProvider found while handling permissions. Was it implemented in the holderImpl?")
			return ctx.Status(500).SendString("Internal Server Error. \nNo AuthProvider found. Please contact the administrator.")
		}

		user := authProvider.User(ctx)
		if user == nil {
			return ctx.Status(401).SendString("Unauthorized")
		}

		for _, perm := range perms {
			if !user.HasPermission(perm) {
				return ctx.Status(403).SendString("Forbidden")
			}
		}

		return ctx.Next()
	}
}
