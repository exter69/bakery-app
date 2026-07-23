package upload

import (
	"context"
	"fmt"
	"io"
	"log"
)

// S3Storage uploads files to an S3-compatible object store (AWS S3, R2, Backblaze B2).
// For now this is a placeholder — local storage is used during development.
type S3Storage struct {
	Bucket  string
	Region  string
	CDNBase string // optional CDN URL prefix (e.g., "https://cdn.example.com")
}

// Upload stores the file in S3 and returns the public URL.
// This is a placeholder implementation that logs but does not actually upload.
func (s *S3Storage) Upload(_ context.Context, key string, _ io.Reader, contentType string) (string, error) {
	log.Printf("[S3Storage] placeholder: would upload %s (type=%s) to bucket=%s region=%s", key, contentType, s.Bucket, s.Region)

	if s.CDNBase != "" {
		return fmt.Sprintf("%s/%s", s.CDNBase, key), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.Bucket, s.Region, key), nil
}
