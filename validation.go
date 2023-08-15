package configreader

import (
	"context"
	"encoding/json"

	"github.com/qri-io/jsonschema"
)

// -----------------------------------------------------------------------------

// ValidationError represents a set of ValidationErrorFailure.
type ValidationError struct {
	Failures []ValidationErrorFailure
}

// ValidationErrorFailure represents a specific JSON schema validation error.
type ValidationErrorFailure struct {
	Location string
	Message  string
}

// -----------------------------------------------------------------------------

func (e *ValidationError) Error() string {
	desc := "validation failed"
	if len(e.Failures) > 0 {
		desc = " / " + e.Failures[0].Message + " @ " + e.Failures[0].Location
	}
	return desc
}

func (*ValidationError) Unwrap() error {
	return nil
}

// -----------------------------------------------------------------------------

func (cr *ConfigReader[T]) validate(encodedSettings []byte) error {
	// Validate against a schema if one is provided
	if len(cr.schema) > 0 {
		schema := []byte(cr.schema)

		// Remove comments from schema
		removeComments(schema)

		// Decode it
		rs := jsonschema.Schema{}
		err := json.Unmarshal(schema, &rs)
		if err != nil {
			return err
		}

		// Execute validation
		var schemaErrors []jsonschema.KeyError

		schemaErrors, err = rs.ValidateBytes(context.Background(), encodedSettings)
		if err != nil {
			return err
		} else if len(schemaErrors) > 0 {
			return newValidationError(schemaErrors)
		}
	}

	// Done
	return nil
}

func newValidationError(errors []jsonschema.KeyError) error {
	err := &ValidationError{
		Failures: make([]ValidationErrorFailure, len(errors)),
	}

	for idx, e := range errors {
		err.Failures[idx].Location = e.PropertyPath
		err.Failures[idx].Message = e.Message
	}

	return err
}
