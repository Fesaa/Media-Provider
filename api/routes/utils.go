package routes

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func ParamsUInt(c *fiber.Ctx, key string, defaultValue ...uint) (uint, error) {
	value, err := strconv.ParseUint(c.Params(key), 10, 32)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return 0, fmt.Errorf("failed to convert: %w", err)
	}

	return uint(value), nil
}

func intQuery(c *fiber.Ctx, key string, def ...int) (int, error) {
	val := c.Query(key)
	if val == "" {

		if len(def) > 0 {
			return def[0], nil
		}

		return 0, errors.New("key not found")
	}

	i, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return i, nil
}
