package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lucatorrekens/bakery-app/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorage implements upload.Storage for testing.
type mockStorage struct {
	lastKey         string
	lastContentType string
	returnURL       string
	returnErr       error
}

func (m *mockStorage) Upload(_ context.Context, key string, _ io.Reader, contentType string) (string, error) {
	m.lastKey = key
	m.lastContentType = contentType
	if m.returnErr != nil {
		return "", m.returnErr
	}
	return m.returnURL, nil
}

// newUploadRequest builds a multipart request with optional file and type field.
func newUploadRequest(t *testing.T, fileName, contentType, uploadType string, body []byte) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if uploadType != "" {
		require.NoError(t, writer.WriteField("type", uploadType))
	}

	if body != nil {
		part, err := writer.CreateFormFile("file", fileName)
		require.NoError(t, err)

		// Write a custom content type header via CreatePart if needed
		_, err = part.Write(body)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/uploads", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// newUploadRequestWithContentType builds a multipart request with a specific content type on the file part.
func newUploadRequestWithContentType(t *testing.T, fileName, contentType, uploadType string, body []byte) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if uploadType != "" {
		require.NoError(t, writer.WriteField("type", uploadType))
	}

	if body != nil {
		h := make(map[string][]string)
		h["Content-Disposition"] = []string{fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName)}
		h["Content-Type"] = []string{contentType}

		part, err := writer.CreatePart(h)
		require.NoError(t, err)
		_, err = part.Write(body)
		require.NoError(t, err)
	}

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/uploads", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// withSellerContext attaches a seller user context to the request.
func withSellerContext(req *http.Request, userID string) *http.Request {
	ctx := middleware.WithUserID(req.Context(), userID)
	ctx = middleware.WithUserRole(ctx, 1) // seller
	return req.WithContext(ctx)
}

func setupUploadRouter(storage *mockStorage) *chi.Mux {
	r := chi.NewRouter()
	handler := NewUploadHandler(storage)
	handler.RegisterRoutes(r)
	return r
}

func TestUpload_ReturnsCreatedWithURL(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/products/abc.jpg"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100) // tiny JPEG-like payload
	req := newUploadRequestWithContentType(t, "photo.jpg", "image/jpeg", "products", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "/uploads/products/abc.jpg", resp["url"])
	assert.Contains(t, storage.lastKey, "products/")
	assert.Equal(t, "image/jpeg", storage.lastContentType)
}

func TestUpload_RejectsMissingTypeField(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/test.jpg"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.jpg", "image/jpeg", "", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TYPE")
}

func TestUpload_RejectsInvalidType(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/test.jpg"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.jpg", "image/jpeg", "avatars", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TYPE")
}

func TestUpload_RejectsInvalidContentType(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/test.jpg"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "doc.pdf", "application/pdf", "products", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_FILE_TYPE")
}

func TestUpload_RejectsMissingFile(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/test.jpg"}
	r := setupUploadRouter(storage)

	// Create request with type but no file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	require.NoError(t, writer.WriteField("type", "products"))
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/uploads", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_FILE")
}

func TestUpload_RejectsUnauthenticatedUser(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/test.jpg"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.jpg", "image/jpeg", "products", body)
	// No context — unauthenticated

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpload_RejectsCustomerRole(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/test.jpg"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.jpg", "image/jpeg", "products", body)
	ctx := middleware.WithUserID(req.Context(), "customer-1")
	ctx = middleware.WithUserRole(ctx, 2) // customer role
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpload_ReturnsErrorOnStorageFailure(t *testing.T) {
	storage := &mockStorage{returnErr: fmt.Errorf("disk full")}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.jpg", "image/jpeg", "products", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "UPLOAD_FAILED")
}

func TestUpload_AcceptsWebPContentType(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/bakeries/abc.webp"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.webp", "image/webp", "bakeries", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, storage.lastKey, "bakeries/")
	assert.Equal(t, "image/webp", storage.lastContentType)
}

func TestUpload_AcceptsPNGContentType(t *testing.T) {
	storage := &mockStorage{returnURL: "/uploads/products/abc.png"}
	r := setupUploadRouter(storage)

	body := bytes.Repeat([]byte{0xFF}, 100)
	req := newUploadRequestWithContentType(t, "photo.png", "image/png", "products", body)
	req = withSellerContext(req, "seller-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, storage.lastKey, "products/")
	assert.Equal(t, "image/png", storage.lastContentType)
}
