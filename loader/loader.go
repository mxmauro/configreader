package loader

import (
	"context"
)

// -----------------------------------------------------------------------------

// Loader defines the spec of a data loader.
type Loader interface {
	Load(ctx context.Context) (data []byte, err error)
}
