package routes

import (
	"fmt"
	"reflect"

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
	Convertor    convertor[T]
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

type convertor[T any] func(ctx *fiber.Ctx) (T, error)

func withDefaultConvertor[T any](kind reflect.Kind) paramOption[T] {
	return func(p *param[T]) {
		var empty T
		convFunc := utils.MustReturn(utils.GetConvertor[T](kind))

		p.Convertor = func(c *fiber.Ctx) (T, error) {
			paramValue := p.GetValue(c)
			if paramValue == "" && !p.AllowEmpty {
				return empty, BadRequest(fmt.Errorf("required parameter %s not present", p.Name))
			}

			value, err := convert(paramValue, p.DefaultValue, convFunc)
			if err != nil {
				return empty, BadRequest(err)
			}

			return value, nil
		}
	}
}

func newParam[T any](name string, options ...paramOption[T]) param[T] {
	p := param[T]{Name: name}

	for _, option := range options {
		option(&p)
	}

	if p.Convertor == nil {
		var empty T
		withDefaultConvertor[T](reflect.ValueOf(empty).Kind())(&p)
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

// withStructConvertor registers a custom convertor. Reflection based if no impl is provided.
// You need to add withAllowEmpty **before** withStructConvertor if the fields are allowed to be empty
// Structs may set the query tag to customize which query param is used
func withStructConvertor[T any](f ...func(ctx *fiber.Ctx) (T, error)) paramOption[T] {
	return func(p *param[T]) {
		if len(f) > 0 {
			p.Convertor = f[0]
			return
		}

		var zero T
		obj := reflect.ValueOf(zero)
		if obj.Type().Kind() == reflect.Ptr {
			obj = obj.Elem()
		}

		if obj.Type().Kind() != reflect.Struct {
			panic(fmt.Sprintf("expected struct, got %T", zero))
		}

		var params []param[any]
		for i := range obj.NumField() {
			opts := []paramOption[any]{withDefaultConvertor[any](obj.Field(i).Kind())}
			if p.AllowEmpty {
				fieldType := obj.Field(i).Type()
				opts = append(opts, withAllowEmpty[any](reflect.Zero(fieldType)))
			}

			fieldName := utils.NonEmpty(obj.Type().Field(i).Tag.Get("query"), obj.Type().Field(i).Name)
			fieldParam := newQueryParam[any](fieldName, opts...)
			params = append(params, fieldParam)
		}

		p.Convertor = func(c *fiber.Ctx) (T, error) {
			var t T

			tVal := reflect.ValueOf(&t).Elem()
			for i, fieldParam := range params {
				v, err := fieldParam.Convertor(c)
				if err != nil {
					return t, BadRequest(err)
				}

				field := tVal.Field(i)
				if field.CanSet() {
					if rv, ok := v.(reflect.Value); ok {
						field.Set(rv)
					} else {
						field.Set(reflect.ValueOf(v))
					}
				} else {
					return t, InternalError(fmt.Errorf("cannot set field %d", i))
				}
			}

			return t, nil
		}
	}
}

func newIdQueryParam() param[int] {
	return newQueryParam[int]("id")
}

func newIdPathParam() param[int] {
	return newPathParam[int]("id")
}

func withParam[T any](param param[T], handler func(*fiber.Ctx, T) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		value, err := param.Convertor(c)
		if err != nil {
			return err
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

	return func(c *fiber.Ctx) error {
		value1, err := param1.Convertor(c)
		if err != nil {
			return err
		}

		value2, err := param2.Convertor(c)
		if err != nil {
			return err
		}

		return handler(c, value1, value2)
	}
}

func withParam3[T1, T2, T3 any](
	param1 param[T1], param2 param[T2], param3 param[T3],
	handler func(*fiber.Ctx, T1, T2, T3) error) fiber.Handler {

	return func(c *fiber.Ctx) error {
		value1, err := param1.Convertor(c)
		if err != nil {
			return err
		}

		value2, err := param2.Convertor(c)
		if err != nil {
			return err
		}

		value3, err := param3.Convertor(c)
		if err != nil {
			return err
		}

		return handler(c, value1, value2, value3)
	}
}
