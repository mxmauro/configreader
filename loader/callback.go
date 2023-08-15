package loader

import (
	"context"
)

// -----------------------------------------------------------------------------

// Callback wraps content to be loaded from a callback function
type Callback struct {
	callback CallbackFunc
}

type CallbackFunc func(ctx context.Context) ([]byte, error)

// -----------------------------------------------------------------------------

// NewCallback create a new callback loader
func NewCallback() *Callback {
	return &Callback{}
}

// WithCallback sets the callback function
func (l *Callback) WithCallback(callback CallbackFunc) *Callback {
	l.callback = callback
	return l
}

// Load loads the content from the callback
func (l *Callback) Load(ctx context.Context) ([]byte, error) {
	return l.callback(ctx)
}
