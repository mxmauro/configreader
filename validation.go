package configreader

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
)

// -----------------------------------------------------------------------------

func (cr *ConfigReader[T]) validate(ctx context.Context, settings *T) error {
	// Execute validation
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.StructCtx(ctx, settings)
	if err != nil {
		var validationErr validator.ValidationErrors

		if errors.As(err, &validationErr) {
			return newValidationError(validationErr)
		}
		return err
	}

	// Execute the extended validation if one was specified
	if cr.extendedValidator != nil {
		err = cr.extendedValidator(ctx, settings)
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
