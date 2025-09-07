package routes

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/Fesaa/Media-Provider/services"
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/gofiber/fiber/v2"
)

type Error struct {
	Err        error
	Caller     string
	StatusCode int
	Extra      fiber.Map
}

func newError(code int, err error, extra ...fiber.Map) *Error {
	return &Error{
		Err:        err,
		StatusCode: code,
		Caller:     getCaller(2),
		Extra:      utils.OrDefault(extra, nil),
	}
}

func newErrorWithDepth(code int, err error, depth int, extra ...fiber.Map) *Error { //nolint: unparam
	return &Error{
		Err:        err,
		StatusCode: code,
		Caller:     getCaller(depth),
		Extra:      utils.OrDefault(extra, nil),
	}
}

func BadRequest(err ...error) *Error {
	return newErrorWithDepth(fiber.StatusBadRequest, utils.OrDefault(err, nil), 3)
}

func InternalError(err error, extra ...fiber.Map) *Error {
	return newErrorWithDepth(fiber.StatusInternalServerError, err, 3, extra...)
}

func NotFound(err ...error) *Error {
	return newErrorWithDepth(fiber.StatusNotFound, utils.OrDefault(err, nil), 3)
}

func Forbidden(err ...error) *Error {
	return newErrorWithDepth(fiber.StatusForbidden, utils.OrDefault(err, nil), 3)
}

func getCaller(depth int) string {
	pc, file, line, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn != nil {
		return fmt.Sprintf("%s:%d (%s)", file, line, fn.Name())
	}

	return fmt.Sprintf("%s:%d", file, line)
}

func (e Error) Error() string {
	if e.Err == nil {
		return "An unknown error occurred"
	}
	return e.Err.Error()
}

var (
	successKey = "success"
	messageKey = "message"
	callerKey  = "caller"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	var e *Error
	if !errors.As(err, &e) {
		return fiber.DefaultErrorHandler(c, err)
	}

	res := fiber.Map{
		successKey: false,
		messageKey: err.Error(),
	}

	_, ok := services.GetFromContextSafe(c, services.UserKey)
	if ok {
		// Don't include caller for non-authenticated endpoints
		res[callerKey] = e.Caller
	}

	if e.Extra != nil {
		for k, v := range e.Extra {
			res[k] = v
		}
	}

	return c.Status(e.StatusCode).JSON(res)
}
