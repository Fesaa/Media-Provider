package routes

import (
	"github.com/Fesaa/Media-Provider/http/payload"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type metadataRoutes struct {
	dig.In

	Router          fiber.Router
	Auth            services.AuthService
	MetadataService services.MetadataService
}

func RegisterMetadataRoutes(mr metadataRoutes) {
	mr.Router.Group("/metadata", mr.Auth.Middleware).
		Get("/", mr.Get)
	// group.Post("/", mr.Update)
}

func (mr *metadataRoutes) Get(c *fiber.Ctx) error {
	log := services.GetFromContext(c, services.LoggerKey)

	m, err := mr.MetadataService.Get()
	if err != nil {
		log.Error().Err(err).Msg("failed to retrieve metadata")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(m)
}

func (mr *metadataRoutes) Update(c *fiber.Ctx, m payload.Metadata) error {
	log := services.GetFromContext(c, services.LoggerKey)

	if err := mr.MetadataService.Update(m); err != nil {
		log.Error().Err(err).Msg("failed to update metadata")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{})
}
