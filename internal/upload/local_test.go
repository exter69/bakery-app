package upload

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorage_UploadCreatesFileAndReturnsURL(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewLocalStorage(dir, "/uploads")
	require.NoError(t, err)

	content := "fake image data"
	url, err := storage.Upload(context.Background(), "products/test.jpg", strings.NewReader(content), "image/jpeg")
	require.NoError(t, err)

	assert.Equal(t, "/uploads/products/test.jpg", url)

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(dir, "products", "test.jpg"))
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestLocalStorage_UploadCreatesSubdirectories(t *testing.T) {
	dir := t.TempDir()
	storage, err := NewLocalStorage(dir, "/uploads")
	require.NoError(t, err)

	_, err = storage.Upload(context.Background(), "bakeries/deep/nested.png", strings.NewReader("data"), "image/png")
	require.NoError(t, err)

	// Verify nested directory was created
	_, err = os.Stat(filepath.Join(dir, "bakeries", "deep", "nested.png"))
	assert.NoError(t, err)
}

func TestNewLocalStorage_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newdir", "uploads")
	_, err := NewLocalStorage(dir, "/uploads")
	require.NoError(t, err)

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
