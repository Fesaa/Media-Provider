package services

import (
	"github.com/Fesaa/Media-Provider/db/models"
	"github.com/go-playground/validator/v10"
	"reflect"
)

func isValidProvider(fl validator.FieldLevel) bool {
	provider := models.Provider(fl.Field().Int())
	return provider >= models.NYAA && provider <= models.BATO
}

func diffValidator(fl validator.FieldLevel) bool {
	currentValue := fl.Field().Interface()

	param := fl.Param()
	parent := fl.Parent().Interface()

	val := reflect.ValueOf(parent).FieldByName(param)
	if !val.IsValid() {
		return false
	}

	return currentValue != val.Interface()
}
