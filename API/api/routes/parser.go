package routes

import (
	"fmt"

	"github.com/Fesaa/Media-Provider/internal/contextkey"
	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
)

// withBody parser the body in the request for you. Keep in mind that this is a terminal operation, no further next
// are called
func withBody[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := contextkey.GetFromContext(c, contextkey.Logger)

		var body T
		if err := c.BodyParser(&body); err != nil {
			log.Error().Err(err).Str("path", c.Path()).Msg("Failed to parse body")
			return BadRequest(err)
		}

		return handler(c, body)
	}
}

// withBodyValidation behaves like withBody, but also runs the body through the validator
func withBodyValidation[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := contextkey.GetFromContext(c, contextkey.Logger)
		serviceProvider := contextkey.GetFromContext(c, contextkey.ServiceProvider)
		validator := utils.MustInvoke[services.ValidationService](serviceProvider)

		var body T
		if err := validator.ValidateCtx(c, &body); err != nil {
			log.Error().Err(err).Str("path", c.Path()).Msg("Failed to validate body")
			return BadRequest(err)
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

func newIdQueryParam() param[int] {
	return newQueryParam[int]("id")
}

func newIdPathParam() param[int] {
	return newPathParam[int]("id")
}

func withParam[T any](param param[T], handler func(*fiber.Ctx, T) error) fiber.Handler {
	convFunc := utils.MustReturn(utils.GetConvertor[T]())

	return func(c *fiber.Ctx) error {
		paramValue := param.GetValue(c)
		if paramValue == "" && !param.AllowEmpty {
			return BadRequest(fmt.Errorf("required parameter %s not present", param.Name))
		}

		value, err := convert(paramValue, param.DefaultValue, convFunc)
		if err != nil {
			return BadRequest(err)
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
	convFunc1 := utils.MustReturn(utils.GetConvertor[T1]())
	convFunc2 := utils.MustReturn(utils.GetConvertor[T2]())

	return func(c *fiber.Ctx) error {
		paramValue1 := param1.GetValue(c)
		if paramValue1 == "" && !param1.AllowEmpty {
			return BadRequest(fmt.Errorf("required parameter %s not present", param1.Name))
		}

		paramValue2 := param2.GetValue(c)
		if paramValue2 == "" && !param2.AllowEmpty {
			return BadRequest(fmt.Errorf("required parameter %s not present", param2.Name))
		}

		value1, err := convert(paramValue1, param1.DefaultValue, convFunc1)
		if err != nil {
			return BadRequest(err)
		}

		value2, err := convert(paramValue2, param2.DefaultValue, convFunc2)
		if err != nil {
			return BadRequest(err)
		}

		return handler(c, value1, value2)
	}
}

func withParam3[T1, T2, T3 any](
	param1 param[T1], param2 param[T2], param3 param[T3],
	handler func(*fiber.Ctx, T1, T2, T3) error) fiber.Handler {
	convFunc1 := utils.MustReturn(utils.GetConvertor[T1]())
	convFunc2 := utils.MustReturn(utils.GetConvertor[T2]())
	convFunc3 := utils.MustReturn(utils.GetConvertor[T3]())

	return func(c *fiber.Ctx) error {
		paramValue1 := param1.GetValue(c)
		if paramValue1 == "" && !param1.AllowEmpty {
			return BadRequest(fmt.Errorf("required parameter %s not present", param1.Name))
		}

		paramValue2 := param2.GetValue(c)
		if paramValue2 == "" && !param2.AllowEmpty {
			return BadRequest(fmt.Errorf("required parameter %s not present", param2.Name))
		}

		paramValue3 := param3.GetValue(c)
		if paramValue3 == "" && !param3.AllowEmpty {
			return BadRequest(fmt.Errorf("required parameter %s not present", param2.Name))
		}

		value1, err := convert(paramValue1, param1.DefaultValue, convFunc1)
		if err != nil {
			return BadRequest(err)
		}

		value2, err := convert(paramValue2, param2.DefaultValue, convFunc2)
		if err != nil {
			return BadRequest(err)
		}

		value3, err := convert(paramValue3, param3.DefaultValue, convFunc3)
		if err != nil {
			return BadRequest(err)
		}

		return handler(c, value1, value2, value3)
	}
}
