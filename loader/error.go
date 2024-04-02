package loader

import (
	"context"

	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

type errorLoader struct {
	err error
}

// -----------------------------------------------------------------------------

func (l *errorLoader) Load(_ context.Context) (model.Values, error) {
	return nil, l.err
}
