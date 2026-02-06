package validator

import (
	"snowflake/pkg/logger"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error       bool
	FailedField string
	Tag         string
	Value       interface{}
}

type Validator struct {
	validator *validator.Validate
	log       *logger.Logger
}

var ValidatorInstance *Validator

func InitValidator(log *logger.Logger) {
	validate := validator.New()

	ValidatorInstance = &Validator{validator: validate, log: log}
}

func (v *Validator) Validate(data interface{}) []ErrorResponse {
	validationErrors := []ErrorResponse{}

	errs := v.validator.Struct(data)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ErrorResponse{
				Error:       true,
				FailedField: err.Field(),
				Tag:         err.Tag(),
				Value:       err.Value(),
			})
		}
	}

	return validationErrors
}
