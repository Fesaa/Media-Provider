package routes

import (
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
)

var features = make([]string, 0)

func init() {
	for _, feature := range utils.SearchOptionFeaturs {
		if utils.FeatureEnabled(feature) {
			features = append(features, feature)
		}
	}
}

func EnabledFeatures(ctx *fiber.Ctx) error {
	return ctx.JSON(features)
}
