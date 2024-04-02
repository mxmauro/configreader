package configreader

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// -----------------------------------------------------------------------------

// ValidationError represents a set of ValidationErrorFailure.
type ValidationError struct {
	Failures []ValidationErrorFailure
}

// ValidationErrorFailure represents a specific JSON schema validation error.
type ValidationErrorFailure struct {
	Field string
	Tag   string
}

// -----------------------------------------------------------------------------

func (e *ValidationError) Error() string {
	sb := strings.Builder{}
	_, _ = sb.WriteString("validation failed")
	for idx := range e.Failures {
		_, _ = sb.WriteString(fmt.Sprintf(" [%d:%s]", idx+1, e.Failures[idx].Error()))
	}
	return sb.String()
}

func (*ValidationError) Unwrap() error {
	return nil
}

func (e *ValidationErrorFailure) Error() string {
	return fmt.Sprintf("unable to validate '%s' on field '%s'", e.Tag, e.Field)
}

// -----------------------------------------------------------------------------

func (cr *ConfigReader[T]) validate(settings *T) error {
	// Execute validation
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(settings)
	if err != nil {
		var validationErr validator.ValidationErrors

		if errors.As(err, &validationErr) {
			return newValidationError(validationErr)
		}
		return err
	}

	// Execute the extended validation if one was specified
	if cr.extendedValidator != nil {
		err = cr.extendedValidator(settings)
		if err != nil {
			return err
		}
	}

	// Done
	return nil
}

// -----------------------------------------------------------------------------

func newValidationError(validationErr validator.ValidationErrors) error {
	err := &ValidationError{
		Failures: make([]ValidationErrorFailure, len(validationErr)),
	}

	for idx, e := range validationErr {
		err.Failures[idx].Field = e.Field()
		err.Failures[idx].Tag = e.Tag()
	}

	return err
}
