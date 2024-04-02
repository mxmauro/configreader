package configreader

import (
	"fmt"
)

// -----------------------------------------------------------------------------

type configLoadError struct {
	innerErr error
}

// -----------------------------------------------------------------------------

func newConfigLoadError(err error) error {
	return &configLoadError{
		innerErr: err,
	}
}

func (e *configLoadError) Error() string {
	return fmt.Sprintf("unable to load configuration [%v]", e.innerErr)
}

func (e *configLoadError) Unwrap() error {
	return e.innerErr
}
