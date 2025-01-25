package services

import (
	"github.com/Fesaa/Media-Provider/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type ValidationService interface {
	Validate(out any) error
	ValidateCtx(ctx *fiber.Ctx, out any) error
}

type validationService struct {
	validator *validator.Validate
	log       zerolog.Logger
}

func ValidatorProvider() *validator.Validate {
	val := validator.New()
	utils.Must(val.RegisterValidation("provider", isValidProvider))
	utils.Must(val.RegisterValidation("diff", diffValidator))
	return val
}

func ValidationServiceProvider(val *validator.Validate, log zerolog.Logger) ValidationService {
	return &validationService{
		validator: val,
		log:       log,
	}
}

func (v *validationService) Validate(out any) error {
	return v.validator.Struct(out)
}

func (v *validationService) ValidateCtx(ctx *fiber.Ctx, out any) error {
	if err := ctx.BodyParser(out); err != nil {
		return err
	}

	return v.Validate(out)
}
