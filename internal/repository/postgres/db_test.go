package postgres

import (
	"context"
	"testing"
	"time"
)

func TestNewPool_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewPool(ctx, Config{DatabaseURL: "not-a-valid-url"})
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}
