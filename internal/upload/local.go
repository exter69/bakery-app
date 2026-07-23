package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage stores uploaded files on the local filesystem.
// Suitable for development; files are served via a static file handler.
type LocalStorage struct {
	Dir     string // directory to store files (e.g., "./uploads")
	BaseURL string // URL prefix for serving (e.g., "/uploads")
}

// NewLocalStorage creates a LocalStorage instance, ensuring the directory exists.
func NewLocalStorage(dir, baseURL string) (*LocalStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("upload: create directory %s: %w", dir, err)
	}
	return &LocalStorage{Dir: dir, BaseURL: baseURL}, nil
}

// Upload writes the file to disk and returns the URL path.
func (s *LocalStorage) Upload(_ context.Context, key string, reader io.Reader, _ string) (string, error) {
	fullPath := filepath.Join(s.Dir, key)

	// Ensure subdirectory exists (e.g., "products/" or "bakeries/")
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("upload: create subdirectory: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("upload: create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("upload: write file: %w", err)
	}

	return s.BaseURL + "/" + key, nil
}
