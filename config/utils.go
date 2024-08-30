package config

import "github.com/gofiber/fiber/v2"

func OrDefault(value string, defaultValue ...string) string {
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return value
}

func Get(ctx *fiber.Ctx) *Config {
	return ctx.Locals("cfg").(*Config)
}
