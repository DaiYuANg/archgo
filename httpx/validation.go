package httpx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/samber/lo"
)

// validateInput validates a typed input value when a validator is configured.
func (s *Server) validateInput(input any) error {
	if s == nil || s.validator == nil || input == nil {
		return nil
	}

	value := reflect.ValueOf(input)
	for value.IsValid() && value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if !value.IsValid() || value.Kind() != reflect.Struct {
		return nil
	}

	return s.validator.Struct(input)
}

// validationErrorMessage converts validator errors into a concise HTTP-facing message.
func validationErrorMessage(err error) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return "request validation failed"
	}

	issues := lo.Map(validationErrs, func(validationErr validator.FieldError, _ int) string {
		field := validationErr.Field()
		if field == "" {
			field = validationErr.StructField()
		}
		if field == "" {
			field = "input"
		}

		return field + " failed '" + validationErr.Tag() + "'"
	})

	if len(issues) == 0 {
		return "request validation failed"
	}

	return "request validation failed: " + strings.Join(issues, "; ")
}
