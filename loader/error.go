package loader

import (
	"context"
)

// -----------------------------------------------------------------------------

type errorLoader struct {
	err error
}

// -----------------------------------------------------------------------------

func (l *errorLoader) Load(_ context.Context) (data []byte, err error) {
	return nil, l.err
}
