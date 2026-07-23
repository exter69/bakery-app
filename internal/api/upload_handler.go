package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lucatorrekens/bakery-app/internal/api/dto"
	"github.com/lucatorrekens/bakery-app/internal/upload"
)

const (
	maxUploadSize = 5 << 20 // 5 MB
)

// allowedContentTypes is the set of accepted image MIME types.
var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// allowedUploadTypes restricts the "type" form field to known prefixes.
var allowedUploadTypes = map[string]bool{
	"products": true,
	"bakeries": true,
}

// UploadHandler handles image upload requests.
type UploadHandler struct {
	storage upload.Storage
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(storage upload.Storage) *UploadHandler {
	return &UploadHandler{storage: storage}
}

// RegisterRoutes registers upload routes on the given chi router.
// The router should already be wrapped with JWT auth middleware.
func (h *UploadHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/uploads", h.Upload)
}

// Upload handles POST /api/uploads.
// Accepts multipart/form-data with fields "file" and "type".
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Require seller or admin role
	userID := requireSeller(w, r)
	if userID == "" {
		return
	}

	// Limit the parsed multipart form to 5 MB + 1 KB overhead for form fields.
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "FILE_TOO_LARGE",
			Message: fmt.Sprintf("file exceeds maximum size of %d MB", maxUploadSize/(1<<20)),
		})
		return
	}

	// Validate the "type" field
	uploadType := r.FormValue("type")
	if !allowedUploadTypes[uploadType] {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_TYPE",
			Message: "type must be one of: products, bakeries",
		})
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "MISSING_FILE",
			Message: "file field is required",
		})
		return
	}
	defer file.Close()

	// Validate file size (double-check against header)
	if header.Size > maxUploadSize {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "FILE_TOO_LARGE",
			Message: fmt.Sprintf("file exceeds maximum size of %d MB", maxUploadSize/(1<<20)),
		})
		return
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	// Some browsers send charset suffix; strip it
	contentType = strings.Split(contentType, ";")[0]
	contentType = strings.TrimSpace(contentType)

	ext, ok := allowedContentTypes[contentType]
	if !ok {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_FILE_TYPE",
			Message: "allowed file types: JPEG, PNG, WebP",
		})
		return
	}

	// Generate unique key
	key := filepath.Join(uploadType, uuid.New().String()+ext)

	// Upload to storage
	url, err := h.storage.Upload(r.Context(), key, file, contentType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "UPLOAD_FAILED",
			Message: "failed to store file",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"url": url,
	})
}
