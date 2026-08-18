package validation

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init initializes the validator
func Init() {
	validate = validator.New()
}

// GetValidator returns the validator instance
func GetValidator() *validator.Validate {
	if validate == nil {
		Init()
	}
	return validate
}

// ValidateStruct validates a struct
func ValidateStruct(s interface{}) error {
	return GetValidator().Struct(s)
}

// ValidateVar validates a single variable
func ValidateVar(field interface{}, tag string) error {
	return GetValidator().Var(field, tag)
}
