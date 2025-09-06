package routes

import (
	"fmt"
	"runtime"

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

func newErrorWithDepth(code int, err error, depth int, extra ...fiber.Map) *Error {
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
		return ""
	}
	return e.Err.Error()
}
