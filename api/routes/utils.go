package routes

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
)

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
