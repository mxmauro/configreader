package configreader

import (
	"bytes"
	"errors"

	"github.com/santhosh-tekuri/jsonschema"
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
		var rs *jsonschema.Schema

		schema := []byte(cr.schema)

		// Remove comments from schema
		schema = removeComments(schema)

		// Decode it
		compiler := jsonschema.NewCompiler()
		err := compiler.AddResource("schema.json", bytes.NewReader(schema))
		if err != nil {
			return err
		}
		rs, err = compiler.Compile("schema.json")
		if err != nil {
			return err
		}

		// Execute validation
		err = rs.Validate(bytes.NewReader(encodedSettings))
		if err != nil {
			var validationErr *jsonschema.ValidationError

			if errors.As(err, &validationErr) {
				return newValidationError(validationErr)
			}
			return err
		}
	}

	// Done
	return nil
}

func newValidationError(validationErr *jsonschema.ValidationError) error {
	err := &ValidationError{
		Failures: make([]ValidationErrorFailure, len(validationErr.Causes)),
	}

	for idx, e := range validationErr.Causes {
		err.Failures[idx].Location = e.SchemaPtr
		err.Failures[idx].Message = e.Message
	}

	return err
}
