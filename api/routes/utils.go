package routes

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

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

func intParam(ctx *fiber.Ctx, param string, d ...int) (int, error) {
	s := ctx.Params(param)
	i, e := strconv.Atoi(s)
	if e != nil {
		if len(d) > 0 {
			return d[0], nil
		}

		return 0, e
	}

	return i, nil
}
