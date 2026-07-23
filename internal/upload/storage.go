package upload

import (
	"context"
	"io"
)

// Storage is the abstraction for file storage.
type Storage interface {
	// Upload stores a file and returns the public URL.
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
}
