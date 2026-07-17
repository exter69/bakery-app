package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text passes through unchanged",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "script tags and content are stripped",
			input:    `<script>alert('xss')</script>`,
			expected: "",
		},
		{
			name:     "script tags with attributes are stripped",
			input:    `<script type="text/javascript">document.cookie</script>`,
			expected: "",
		},
		{
			name:     "bold tags are stripped leaving content",
			input:    "<b>bold</b>",
			expected: "bold",
		},
		{
			name:     "anchor tags are stripped leaving content",
			input:    `<a href="http://evil.com">click me</a>`,
			expected: "click me",
		},
		{
			name:     "mixed content is properly cleaned",
			input:    `Hello <script>alert('xss')</script> <b>World</b>!`,
			expected: "Hello World!",
		},
		{
			name:     "div tags are stripped leaving content",
			input:    "<div>content inside div</div>",
			expected: "content inside div",
		},
		{
			name:     "nested tags are stripped",
			input:    "<div><p><b>nested</b></p></div>",
			expected: "nested",
		},
		{
			name:     "self-closing tags are stripped",
			input:    `text <br/> more text <img src="x"/>`,
			expected: "text more text",
		},
		{
			name:     "empty string passes through",
			input:    "",
			expected: "",
		},
		{
			name:     "text with angle brackets not forming tags passes through",
			input:    "5 > 3 and 2 < 4",
			expected: "5 > 3 and 2 < 4",
		},
		{
			name:     "case insensitive script removal",
			input:    `<SCRIPT>alert('xss')</SCRIPT>`,
			expected: "",
		},
		{
			name:     "multiline script content is removed",
			input:    "<script>\nvar x = 1;\nalert(x);\n</script>safe text",
			expected: "safe text",
		},
		{
			name:     "multiple script tags are all removed",
			input:    `<script>a()</script>text<script>b()</script>`,
			expected: "text",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizeString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInputSanitizer_Middleware(t *testing.T) {
	// Handler that echoes back the parsed body
	echoHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	})

	handler := InputSanitizer(echoHandler)

	t.Run("sanitizes string fields in JSON body", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":        "<b>Bakery</b>",
			"description": `<script>alert('xss')</script>Nice place`,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var result map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Bakery", result["name"])
		assert.Equal(t, "Nice place", result["description"])
	})

	t.Run("preserves non-string fields", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":     "<b>Test</b>",
			"quantity": 5,
			"price":    9.99,
			"active":   true,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var result map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, "Test", result["name"])
		assert.Equal(t, float64(5), result["quantity"])
		assert.Equal(t, 9.99, result["price"])
		assert.Equal(t, true, result["active"])
	})

	t.Run("sanitizes nested objects", func(t *testing.T) {
		payload := map[string]interface{}{
			"order": map[string]interface{}{
				"note": "<script>steal()</script>Please deliver fast",
			},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var result map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		order := result["order"].(map[string]interface{})
		assert.Equal(t, "Please deliver fast", order["note"])
	})

	t.Run("sanitizes arrays of strings", func(t *testing.T) {
		payload := map[string]interface{}{
			"tags": []interface{}{"<b>fresh</b>", "<script>x</script>organic"},
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var result map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		tags := result["tags"].([]interface{})
		assert.Equal(t, "fresh", tags[0])
		assert.Equal(t, "organic", tags[1])
	})

	t.Run("passes through non-JSON content types", func(t *testing.T) {
		body := []byte("<b>not json</b>")

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, "<b>not json</b>", rec.Body.String())
	})

	t.Run("passes through requests with no body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("handles empty JSON body gracefully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("handles invalid JSON gracefully", func(t *testing.T) {
		body := []byte(`{invalid json`)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Should pass through unchanged
		assert.Equal(t, `{invalid json`, rec.Body.String())
	})
}
