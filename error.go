package configreader

import (
	"fmt"
	"strings"
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
