package api

import (
	"github.com/Fesaa/Media-Provider/api/routes"
	"github.com/Fesaa/Media-Provider/middelware"
	"github.com/Fesaa/Media-Provider/models"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, holder models.Holder) error {
	api := app.Group("/api")

	api.Post("/login", routes.Login)
	api.Get("/logout", routes.Logout)
	api.Post("/register", routes.Register)

	api.Post("/search", routes.Search)

	admin := api.Group("/admin", middleware.HasPermissions(holder, "STAFF"))
	permissions := admin.Group("/permissions")

	permissions.Get("/", middleware.HasPermissions(holder, "GET_PERMS"), routes.GetPerms)
	permissions.Post("/refresh", middleware.HasPermissions(holder, "REFRESH_PERMS"), routes.RefreshPerms)

	return nil
}
