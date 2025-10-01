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
	return withParams(handler, newBodyParam[T]())
}

// withBodyValidation behaves like withBody, but also runs the body through the validator
func withBodyValidation[T any](handler func(*fiber.Ctx, T) error) fiber.Handler {
	return withParams(handler, newValidatedBodyParam[T]())
}

type paramType int

const (
	pathParam paramType = iota
	queryParam
	bodyParam
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

func newBodyParam[T any](options ...paramOption[T]) param[T] {
	options = append(options, func(p *param[T]) {
		p.Type = bodyParam
		p.Convertor = func(c *fiber.Ctx) (T, error) {
			var body T
			if err := c.BodyParser(&body); err != nil {
				return body, BadRequest(err)
			}
			return body, nil
		}
	})

	return newParam[T]("", options...)
}

func newValidatedBodyParam[T any](options ...paramOption[T]) param[T] {
	options = append(options, func(p *param[T]) {
		p.Type = bodyParam
		p.Convertor = func(c *fiber.Ctx) (T, error) {
			var body T
			serviceProvider := contextkey.GetFromContext(c, contextkey.ServiceProvider)
			validator := utils.MustInvoke[services.ValidationService](serviceProvider)
			if err := validator.ValidateCtx(c, &body); err != nil {
				return body, BadRequest(err)
			}
			return body, nil
		}
	})

	return newParam[T]("", options...)
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

func convert[T any](v string, def T, conv func(string) (T, error)) (T, error) {
	if len(v) == 0 {
		return def, nil
	}

	return conv(v)
}

func withParams(handler any, params ...any) fiber.Handler {
	convs := setupConvertors(handler, params...)

	funcValue := reflect.ValueOf(handler)
	if funcValue.IsZero() || !funcValue.IsValid() {
		panic(fmt.Sprintf("handler did not have a function"))
	}

	return func(c *fiber.Ctx) error {
		args := make([]reflect.Value, len(params)+1)

		for i := range len(params) {
			ret := convs[i].Call([]reflect.Value{reflect.ValueOf(c)})
			arg := ret[0].Interface()
			if err, ok := ret[1].Interface().(error); ok && err != nil {
				return fmt.Errorf("param %d conversion failed: %w", i, err)
			}

			if v, ok := arg.(reflect.Value); ok {
				args[i+1] = v
			} else {
				args[i+1] = reflect.ValueOf(arg)
			}
		}

		args[0] = reflect.ValueOf(c)
		ret := funcValue.Call(args)
		err, _ := ret[0].Interface().(error)
		return err
	}
}

func setupConvertors(handler any, params ...any) []reflect.Value {
	f := reflect.TypeOf(handler)
	if f.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected function, got %T", handler))
	}
	if f.NumIn() != len(params)+1 {
		panic(fmt.Sprintf("expected %d args, got %d", len(params)+1, f.NumIn()))
	}
	if f.NumOut() != 1 {
		panic(fmt.Sprintf("expected 1 args, got %d", f.NumOut()))
	}

	if f.In(0) != reflect.TypeOf((*fiber.Ctx)(nil)) {
		panic(fmt.Sprintf("first parameter must be *fiber.Ctx, got %s", f.In(0)))
	}

	if !f.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic(fmt.Sprintf("return type must be error, got %s", f.Out(0)))
	}

	convs := make([]reflect.Value, len(params))
	for i, p := range params {
		conv := reflect.ValueOf(p).FieldByName("Convertor")
		if !conv.IsValid() {
			panic(fmt.Sprintf("param did not have a expected Convertor"))
		}

		if conv.Kind() != reflect.Func {
			panic(fmt.Sprintf("expected function, got %s", conv.Kind().String()))
		}

		convType := conv.Type()
		if convType.NumIn() != 1 {
			panic(fmt.Sprintf("param %d Convertor: expected 1 input, got %d", i, convType.NumIn()))
		}
		if convType.In(0) != reflect.TypeOf((*fiber.Ctx)(nil)) {
			panic(fmt.Sprintf("param %d Convertor: input must be *fiber.Ctx, got %s", i, convType.In(0)))
		}
		if convType.NumOut() != 2 {
			panic(fmt.Sprintf("param %d Convertor: expected 2 outputs, got %d", i, convType.NumOut()))
		}
		if !convType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			panic(fmt.Sprintf("param %d Convertor: second return must be error, got %s", i, convType.Out(1)))
		}

		handlerParamType := f.In(i + 1)
		convertorReturnType := convType.Out(0)

		if convertorReturnType != handlerParamType {
			panic(fmt.Sprintf(
				"param %d type mismatch: handler expects %s, but Convertor returns %s",
				i, handlerParamType, convertorReturnType,
			))
		}

		convs[i] = conv
	}

	return convs
}
