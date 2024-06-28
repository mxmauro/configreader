package parsers

import (
	"errors"
)

// -----------------------------------------------------------------------------

var (
	errInvalidValue = errors.New("invalid value")
	errInvalidUnit  = errors.New("invalid unit")

	errInvalidDuration = errors.New("invalid duration")

	errOverflow = errors.New("overflow")
)
