package routes

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"go.uber.org/dig"
)

type metadataRoutes struct {
	dig.In

	Router          fiber.Router
	Auth            services.AuthService `name:"jwt-auth"`
	MetadataService services.MetadataService
	Log             zerolog.Logger
}

func RegisterMetadataRoutes(mr metadataRoutes) {
	group := mr.Router.Group("/metadata", mr.Auth.Middleware)

	group.Get("/", mr.Get)
	// group.Post("/", mr.Update)
}

func (mr *metadataRoutes) Get(c *fiber.Ctx) error {
	m, err := mr.MetadataService.Get()
	if err != nil {
		mr.Log.Error().Err(err).Msg("failed to retrieve metadata")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(m)
}

func (mr *metadataRoutes) Update(c *fiber.Ctx, m payload.Metadata) error {
	if err := mr.MetadataService.Update(m); err != nil {
		mr.Log.Error().Err(err).Msg("failed to update metadata")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{})
}
