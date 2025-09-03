package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// WithBody parser the body in the request for you. Keep in mind that this is a terminal operation, no further next
// are called
func WithBody[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body T
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, body)
	}
}

// WithBodyValidation behaves like WithBody, but also runs the body through the validator
func WithBodyValidation[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		serviceProvider := c.Locals(services.ServiceProviderKey).(*dig.Container)
		validator := utils.MustInvoke[services.ValidationService](serviceProvider)

		var body T
		if err := validator.ValidateCtx(c, body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, body)
	}
}

type ParamsOptions[T any] struct {
	Name         string
	AllowEmpty   bool
	DefaultValue T
	Message      string
}

func WithQueryName[T any](s string) ParamsOptions[T] {
	return ParamsOptions[T]{
		Name:       s,
		AllowEmpty: false,
	}
}

func WithAllowEmpty[T any](s string, defs ...T) ParamsOptions[T] {
	def := func() T {
		if len(defs) == 0 {
			var zero T
			return zero
		}

		return defs[0]
	}()

	return ParamsOptions[T]{
		Name:         s,
		AllowEmpty:   true,
		DefaultValue: def,
	}
}

func WithMessage[T any](s string, message string) ParamsOptions[T] {
	return ParamsOptions[T]{
		Name:    s,
		Message: message,
	}
}

func IdParamsOption() ParamsOptions[uint] {
	return ParamsOptions[uint]{
		Name: "id",
	}
}

// WithQueryParams converts the query params based on the type
func WithQueryParams[T any](options ParamsOptions[T], handler func(*fiber.Ctx, T) error) fiber.Handler {
	convFunc := utils.MustReturn(getConvertor[T]())

	return func(c *fiber.Ctx) error {
		queryParam := c.Query(options.Name)
		if queryParam == "" {
			if options.AllowEmpty {
				return handler(c, options.DefaultValue)
			}

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter is empty",
			})
		}

		value, err := convFunc(queryParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return handler(c, value)
	}
}

// WithParams converts the params based on the type
func WithParams[T any](options ParamsOptions[T], handler func(*fiber.Ctx, T) error) fiber.Handler {
	convFunc := utils.MustReturn(getConvertor[T]())

	return func(c *fiber.Ctx) error {
		param := c.Params(options.Name)
		if param == "" {
			if options.AllowEmpty {
				return handler(c, options.DefaultValue)
			}

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "query parameter is empty",
			})
		}

		value, err := convFunc(param)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": options.Message,
			})
		}

		return handler(c, value)
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
