package routes

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
)

// withBody parser the body in the request for you. Keep in mind that this is a terminal operation, no further next
// are called
func withBody[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := services.GetFromContext(c, services.LoggerKey)

		var body T
		if err := c.BodyParser(&body); err != nil {
			log.Error().Err(err).Str("path", c.Path()).Msg("Failed to parse body")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, body)
	}
}

// withBodyValidation behaves like withBody, but also runs the body through the validator
func withBodyValidation[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := services.GetFromContext(c, services.LoggerKey)
		serviceProvider := services.GetFromContext(c, services.ServiceProviderKey)
		validator := utils.MustInvoke[services.ValidationService](serviceProvider)

		var body T
		if err := validator.ValidateCtx(c, &body); err != nil {
			log.Error().Err(err).Str("path", c.Path()).Msg("Failed to validate body")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, body)
	}
}

type paramType int

const (
	pathParam paramType = iota
	queryParam
)

type param[T any] struct {
	Name         string
	Type         paramType
	AllowEmpty   bool
	DefaultValue T
	Message      string
}

func (p param[T]) GetValue(ctx *fiber.Ctx) string {
	switch p.Type {
	case pathParam:
		return ctx.Params(p.Name)
	case queryParam:
		return ctx.Query(p.Name)
	}

	return ""
}

type paramOption[T any] func(*param[T])

func newParam[T any](name string, options ...paramOption[T]) param[T] {
	p := param[T]{Name: name}

	for _, option := range options {
		option(&p)
	}

	return p
}

func newPathParam[T any](name string, options ...paramOption[T]) param[T] {
	options = append(options, func(p *param[T]) {
		p.Type = pathParam
	})
	return newParam[T](name, options...)
}

func newQueryParam[T any](name string, options ...paramOption[T]) param[T] {
	options = append(options, func(p *param[T]) {
		p.Type = queryParam
	})

	return newParam[T](name, options...)
}

func withAllowEmpty[T any](defs ...T) paramOption[T] {
	def := func() T {
		if len(defs) == 0 {
			var zero T
			return zero
		}

		return defs[0]
	}()

	return func(p *param[T]) {
		p.AllowEmpty = true
		p.DefaultValue = def
	}
}

func withMessage[T any](msg string) paramOption[T] {
	return func(p *param[T]) {
		p.Message = msg
	}
}

func newIdQueryParam() param[uint] {
	return newQueryParam[uint]("id")
}

func newIdPathParam() param[uint] {
	return newPathParam[uint]("id")
}

func withParam[T any](param param[T], handler func(*fiber.Ctx, T) error) fiber.Handler {
	convFunc := utils.MustReturn(getConvertor[T]())

	return func(c *fiber.Ctx) error {
		paramValue := param.GetValue(c)
		if paramValue == "" {
			if param.AllowEmpty {
				return handler(c, param.DefaultValue)
			}

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter is empty",
			})
		}

		value, err := convFunc(paramValue)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, value)
	}
}

func convert[T any](v string, def T, conv func(string) (T, error)) (T, error) {
	if len(v) == 0 {
		return def, nil
	}

	return conv(v)
}

func withParam2[T1, T2 any](
	param1 param[T1], param2 param[T2],
	handler func(*fiber.Ctx, T1, T2) error) fiber.Handler {
	convFunc1 := utils.MustReturn(getConvertor[T1]())
	convFunc2 := utils.MustReturn(getConvertor[T2]())

	return func(c *fiber.Ctx) error {
		paramValue1 := param1.GetValue(c)
		if paramValue1 == "" && !param1.AllowEmpty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is required", param1.Name),
			})
		}

		paramValue2 := param2.GetValue(c)
		if paramValue2 == "" && !param2.AllowEmpty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is required", param2.Name),
			})
		}

		value1, err := convert(paramValue1, param1.DefaultValue, convFunc1)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		value2, err := convert(paramValue2, param2.DefaultValue, convFunc2)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, value1, value2)
	}
}

func withParam3[T1, T2, T3 any](
	param1 param[T1], param2 param[T2], param3 param[T3],
	handler func(*fiber.Ctx, T1, T2, T3) error) fiber.Handler {
	convFunc1 := utils.MustReturn(getConvertor[T1]())
	convFunc2 := utils.MustReturn(getConvertor[T2]())
	convFunc3 := utils.MustReturn(getConvertor[T3]())

	return func(c *fiber.Ctx) error {
		paramValue1 := param1.GetValue(c)
		if paramValue1 == "" && !param1.AllowEmpty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is required", param1.Name),
			})
		}

		paramValue2 := param2.GetValue(c)
		if paramValue2 == "" && !param2.AllowEmpty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is required", param2.Name),
			})
		}

		paramValue3 := param3.GetValue(c)
		if paramValue3 == "" && !param3.AllowEmpty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is required", param3.Name),
			})
		}

		value1, err := convert(paramValue1, param1.DefaultValue, convFunc1)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		value2, err := convert(paramValue2, param2.DefaultValue, convFunc2)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		value3, err := convert(paramValue3, param3.DefaultValue, convFunc3)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, value1, value2, value3)
	}
}

func getConvertor[T any]() (func(string) (T, error), error) {
	var zero T

	switch any(zero).(type) {
	case int:
		return func(s string) (T, error) {
			val, err := strconv.Atoi(s)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case int64:
		return func(s string) (T, error) {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case int32:
		return func(s string) (T, error) {
			val, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return zero, err
			}
			return any(int32(val)).(T), nil
		}, nil
	case uint:
		return func(s string) (T, error) {
			val, err := strconv.ParseUint(s, 10, 0)
			if err != nil {
				return zero, err
			}
			return any(uint(val)).(T), nil
		}, nil
	case uint64:
		return func(s string) (T, error) {
			val, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case float32:
		return func(s string) (T, error) {
			val, err := strconv.ParseFloat(s, 32)
			if err != nil {
				return zero, err
			}
			return any(float32(val)).(T), nil
		}, nil
	case float64:
		return func(s string) (T, error) {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case bool:
		return func(s string) (T, error) {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return zero, err
			}
			return any(val).(T), nil
		}, nil
	case string:
		return func(s string) (T, error) {
			return any(s).(T), nil
		}, nil
	case time.Time:
		return func(s string) (T, error) {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return zero, err
			}
			return any(t).(T), nil
		}, nil
	}

	return nil, fmt.Errorf("unknown type %T", any(zero))
}
