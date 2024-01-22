package loader

import (
	"context"
	"errors"
	"os"
)

// -----------------------------------------------------------------------------

// Memory wraps content to be loaded from a string
type Memory struct {
	data []byte

	err error
}

// -----------------------------------------------------------------------------

// NewMemory create a new memory data loader
func NewMemory() *Memory {
	return &Memory{}
}

// NewMemoryFromEnvironmentVariable create a new memory data loader from an environment variable
func NewMemoryFromEnvironmentVariable(Name string) *Memory {
	l := &Memory{}

	data, ok := os.LookupEnv(Name)
	if ok {
		if len(data) > 0 {
			// Make a copy of the source data, so we can safely manipulate it
			l.data = []byte(data)
		} else {
			l.err = errors.New("environment variable '" + Name + "' is empty")
		}
	} else {
		l.err = errors.New("environment variable '" + Name + "' not found")
	}

	// Done
	return l
}

// WithData sets the data to return when the content is loaded
func (l *Memory) WithData(data string) *Memory {
	if l.err == nil {
		// Make a copy of the source data, so we can safely manipulate it
		l.data = []byte(data)
	}
	return l
}

// Load loads the content from the file
func (l *Memory) Load(_ context.Context) ([]byte, error) {
	// If an error was set by a With... function, return it
	if l.err != nil {
		return nil, l.err
	}

	// Return content
	return l.data, nil
}
