// Package validator - configuracion central de validador
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator struct
type Validator struct {
	validate *validator.Validate
}

// New validator
func New() *Validator {
	return &Validator{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// ValidationError structs
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate returns []ValidationError
func (v *Validator) Validate(s interface{}) []ValidationError {
	err := v.validate.Struct(s)

	if err == nil {
		return nil
	}

	var errors []ValidationError
	validationErrors, ok := err.(validator.ValidationErrors) // Asercion
	if !ok {
		return nil
	}
	for _, e := range validationErrors {
		errors = append(errors, ValidationError{
			Field:   strings.ToLower(e.Field()),
			Message: translateTag(e),
		})
	}
	return errors
}

func translateTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "este campo es obligatorio"
	case "email":
		return "debe ser un correo válido"
	case "min":
		return fmt.Sprintf("debe tener mínimo %s caracteres", e.Param())
	case "max":
		return fmt.Sprintf("debe tener máximo %s caracteres", e.Param())
	default:
		return "valor inválido"
	}
}
