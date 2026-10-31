// Package validator provides request validation helpers.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator validates values using struct tags.
type Validator struct {
	validate *validator.Validate
}

// New creates a validator.
func New() *Validator {
	return &Validator{
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// ValidationError describes a field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate validates s and returns details for each validation error.
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
